package db

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

type DummyDB[Model any] struct {
	list     func() ([]Model, error)
	retrieve func(id any) (Model, error)
	create   func(m *Model) (Model, error)
	update   func(id any, new *Model) (Model, error)
	delete   func(id any) error

	q *Queries[Model]
}

func (d DummyDB[Model]) Pagination() Pagination {
	return DummyPagination[Model]{}
}

func (d DummyDB[Model]) Filter() QueryMod {
	return dummyQueryMod{}
}

func (d DummyDB[Model]) Order() QueryMod {
	return dummyQueryMod{}
}

func (d DummyDB[Model]) Queries() *Queries[Model] {
	return d.q.WithCreate(func(ctx *gin.Context, m *Model) (Model, error) {
		return d.create(m)
	}).WithUpdate(func(ctx *gin.Context, old *Model, new *Model, id any) (Model, error) {
		return d.update(id, new)
	}).WithDelete(func(ctx *gin.Context, m Model, id any) error {
		return d.delete(id)
	}).WithRetrieve(func(ctx *gin.Context, id any) (models.InternalValue, error) {
		m, err := d.retrieve(id)
		if err != nil {
			return nil, err
		}
		return models.AsInternalValue(m)
	}).WithList(func(ctx *gin.Context) ([]models.InternalValue, error) {
		ms, err := d.list()
		if err != nil {
			return nil, err
		}
		ivs := make([]models.InternalValue, len(ms))
		for i, m := range ms {
			iv, err := models.AsInternalValue(m)
			if err != nil {
				return nil, err
			}
			ivs[i] = iv
		}
		return ivs, nil
	})
}

func (d DummyDB[Model]) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

type DummyPagination[Model any] struct {
}

func (d DummyPagination[Model]) Apply(ctx *gin.Context) {
}

func (d DummyPagination[Model]) Format(ctx *gin.Context, models []any) (any, error) {
	return models, nil
}

type dummyQueryMod struct {
}

func (d dummyQueryMod) Apply(ctx *gin.Context) {
}

func Dummy[Model any](seed ...Model) *DummyDB[Model] {
	return &DummyDB[Model]{
		q: &Queries[Model]{},
		list: func() ([]Model, error) {
			return seed, nil
		},
		retrieve: func(id any) (Model, error) {
			return seed[0], nil
		},
		create: func(m *Model) (Model, error) {
			return seed[0], nil
		},
		update: func(id any, new *Model) (Model, error) {
			return seed[0], nil
		},
		delete: func(id any) error {
			return nil
		},
	}
}
