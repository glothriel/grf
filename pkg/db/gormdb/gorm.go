package gormdb

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"gorm.io/gorm"
)

type GormFilterFunc func(ctx *gin.Context, db *gorm.DB) *gorm.DB

type gormQueryMod[Model any] struct {
	modFunc GormFilterFunc
}

func (g gormQueryMod[Model]) Apply(ctx *gin.Context) {
	CtxSetQuery[Model](ctx, g.modFunc(ctx, CtxQuery[Model](ctx)))
}

type gormPagination[Model any] struct {
	child Pagination
}

func (g gormPagination[Model]) Apply(ctx *gin.Context) {
	CtxSetQuery[Model](ctx, g.child.Apply(ctx, CtxQuery[Model](ctx)))
}

func (g gormPagination[Model]) Format(ctx *gin.Context, elems []any) (any, error) {
	return g.child.Format(ctx, elems)
}

type GormDatabase[Model any] struct {
	queries    *db.Queries[Model]
	filter     *gormQueryMod[Model]
	order      *gormQueryMod[Model]
	pagination *gormPagination[Model]

	middleware []gin.HandlerFunc
}

func (g GormDatabase[Model]) Queries() *db.Queries[Model] {
	return g.queries
}

func (g GormDatabase[Model]) Filter() db.QueryMod {
	return g.filter
}

func (g GormDatabase[Model]) Order() db.QueryMod {
	return g.order
}

func (g GormDatabase[Model]) Pagination() db.Pagination {
	return g.pagination
}

func (g GormDatabase[Model]) Middleware() []gin.HandlerFunc {
	return g.middleware
}

func (g *GormDatabase[Model]) WithFilter(filterFunc GormFilterFunc) *GormDatabase[Model] {
	g.filter.modFunc = filterFunc
	return g
}

func (g *GormDatabase[Model]) WithPagination(pagination Pagination) *GormDatabase[Model] {
	g.pagination.child = pagination
	return g
}

func (g *GormDatabase[Model]) WithOrderBy(orderClause any) *GormDatabase[Model] {
	g.order.modFunc = func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		return db.Order(orderClause)
	}
	return g
}

func Gorm[Model any](db *gorm.DB) *GormDatabase[Model] {
	return &GormDatabase[Model]{
		queries: GormQueries[Model](),
		filter: &gormQueryMod[Model]{
			modFunc: func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				return db
			},
		},
		order: &gormQueryMod[Model]{
			modFunc: func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				return db
			},
		},
		pagination: &gormPagination[Model]{
			child: &NoPagination{},
		},
		middleware: []gin.HandlerFunc{
			func(ctx *gin.Context) {
				CtxSetGorm(ctx, db)
				CtxInitQuery[Model](ctx)
				ctx.Next()
			},
		},
	}
}

// GormQueries returns default queries providing basic CRUD functionality
func GormQueries[Model any]() *db.Queries[Model] {
	converter := FromDBConverter[Model]()
	var emptyModelForDBModel Model

	return &db.Queries[Model]{
		List: func(ctx *gin.Context) ([]models.InternalValue, error) {
			rawEntities := []map[string]any{}
			findErr := CtxQuery[Model](ctx).Model(&emptyModelForDBModel).Find(&rawEntities).Error
			if findErr != nil {
				return nil, findErr
			}
			internalValues := make([]models.InternalValue, len(rawEntities))
			for i, rawEntity := range rawEntities {
				internalValue, convertErr := converter(rawEntity)
				if convertErr != nil {
					return nil, convertErr
				}
				internalValues[i] = internalValue
			}
			return internalValues, findErr
		},
		Retrieve: func(ctx *gin.Context, id any) (models.InternalValue, error) {
			var rawEntity map[string]any
			retrieveErr := CtxQuery[Model](ctx).Model(&emptyModelForDBModel).First(&rawEntity, "id = ?", id).Error
			if retrieveErr != nil {
				return nil, retrieveErr
			}
			return converter(rawEntity)
		},
		Create: func(ctx *gin.Context, m *Model) (Model, error) {
			createErr := CtxQuery[Model](ctx).Create(m).Error
			return *m, createErr
		},
		Update: func(ctx *gin.Context, old *Model, new *Model, id any) (Model, error) {
			updateErr := CtxQuery[Model](ctx).Model(new).Updates(new).Error
			if updateErr != nil {
				return *new, updateErr
			}
			return *new, nil
		},
		Delete: func(ctx *gin.Context, m Model, id any) error {
			deleteErr := CtxQuery[Model](ctx).Delete(&m, "id = ?", id).Error
			return deleteErr
		},
	}
}

// FromDBConverter internally uses *sql.Scanner to convert a map[string]any to an InternalValue
func FromDBConverter[Model any]() func(map[string]any) (models.InternalValue, error) {
	fromDB := fields.SQLScannerOrPassthrough[Model]()
	return func(m map[string]any) (models.InternalValue, error) {
		intVal := models.InternalValue{}
		for k := range m {
			vAsIntVal, fromDBErr := fromDB(m, k, nil)
			if fromDBErr != nil {
				return nil, fromDBErr
			}
			intVal[k] = vAsIntVal
		}
		return intVal, nil
	}
}
