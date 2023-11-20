package dummy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/queries/crud"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InMemoryQueryDriver is a dummy query driver that stores all data in memory.
type InMemoryQueryDriver[Model any] struct {
	list     crud.ListQueryFunc
	create   crud.CreateQueryFunc
	retrieve func(id any) (models.InternalValue, error)
	update   func(id any, new models.InternalValue) (models.InternalValue, error)
	delete   func(id any) error

	q *crud.CRUD[Model]
}

// Pagination implements db.QueryDriver interface
func (d InMemoryQueryDriver[Model]) Pagination() common.Pagination {
	return dummyPagination[Model]{}
}

// Filter implements db.QueryDriver interface
func (d InMemoryQueryDriver[Model]) Filter() common.QueryMod {
	return dummyQueryMod{}
}

// Order implements db.QueryDriver interface
func (d InMemoryQueryDriver[Model]) Order() common.QueryMod {
	return dummyQueryMod{}
}

// CRUD implements db.QueryDriver interface
func (d InMemoryQueryDriver[Model]) CRUD() *crud.CRUD[Model] {
	return d.q.WithCreate(func(ctx *gin.Context, m models.InternalValue) (models.InternalValue, error) {
		return d.create(ctx, m)
	}).WithUpdate(func(
		ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any,
	) (models.InternalValue, error) {
		return d.update(id, new)
	}).WithDestroy(func(ctx *gin.Context, id any) error {
		return d.delete(id)
	}).WithRetrieve(func(ctx *gin.Context, id any) (models.InternalValue, error) {
		return d.retrieve(id)
	}).WithList(func(ctx *gin.Context) ([]models.InternalValue, error) {
		return d.list(ctx)
	})
}

func (d *InMemoryQueryDriver[Model]) WithCreate(f crud.CreateQueryFunc) *InMemoryQueryDriver[Model] {
	d.create = f
	return d
}

// Middleware implements db.QueryDriver interface
func (d InMemoryQueryDriver[Model]) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

type dummyPagination[Model any] struct {
}

func (d dummyPagination[Model]) Apply(*gin.Context) {
}

func (d dummyPagination[Model]) Format(_ *gin.Context, models []any) (any, error) {
	return models, nil
}

type dummyQueryMod struct {
}

func (d dummyQueryMod) Apply(*gin.Context) {
}

// InMemoryDriver creates InMemoryQueryDriver with given seed data.
func InMemoryDriver[Model any](seed ...Model) *InMemoryQueryDriver[Model] {
	storage := map[any]models.InternalValue{}
	var newID = newIDGenerator[Model](storage)
	driver := &InMemoryQueryDriver[Model]{
		q: &crud.CRUD[Model]{},
		list: func(*gin.Context) ([]models.InternalValue, error) {
			ivs := make([]models.InternalValue, len(storage))
			i := 0
			for _, v := range storage {
				ivs[i] = v
				i++
			}
			return ivs, nil
		},
		retrieve: func(id any) (models.InternalValue, error) {
			elem, ok := storage[fmt.Sprintf("%v", id)]
			if !ok {
				return nil, common.ErrorNotFound
			}
			return elem, nil
		},
		create: func(_ *gin.Context, m models.InternalValue) (models.InternalValue, error) {
			m["id"] = newID()
			storage[fmt.Sprintf("%v", m["id"])] = m
			return m, nil
		},
		update: func(id any, m models.InternalValue) (models.InternalValue, error) {
			if _, ok := storage[fmt.Sprintf("%v", id)]; !ok {
				return nil, common.ErrorNotFound
			}
			storage[fmt.Sprintf("%v", id)] = m
			return m, nil
		},
		delete: func(id any) error {
			if _, ok := storage[fmt.Sprintf("%v", id)]; !ok {
				return common.ErrorNotFound
			}
			delete(storage, fmt.Sprintf("%v", id))
			return nil
		},
	}
	for _, m := range seed {
		intVal := models.AsInternalValue(m)
		_, createErr := driver.create(nil, intVal)
		if createErr != nil {
			panic(createErr)
		}
	}

	return driver
}

func newIDGenerator[Model any](storage map[any]models.InternalValue) func() any {
	var currModel Model
	intVal := models.AsInternalValue(currModel)
	if _, ok := intVal["id"]; !ok {
		logrus.Panic("Model needs to have a field that is serialized to 'id'")
	}
	if _, ok := intVal["id"].(uint); ok {
		return func() any {
			return uint(len(storage) + 1)
		}
	} else if _, ok := intVal["id"].(int); ok {
		return func() any {
			return int(len(storage) + 1)
		}
	} else if _, ok := intVal["id"].(string); ok {
		return func() any {
			return uuid.New().String()
		}
	}
	logrus.Panic("id must be uint or string")
	return nil
}
