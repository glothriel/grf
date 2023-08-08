package gorm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/queries/crud"
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

type GormQueryDriver[Model any] struct {
	crud       *crud.CRUD[Model]
	filter     *gormQueryMod[Model]
	order      *gormQueryMod[Model]
	pagination *gormPagination[Model]

	middleware []gin.HandlerFunc
}

func (g GormQueryDriver[Model]) CRUD() *crud.CRUD[Model] {
	return g.crud
}

func (g GormQueryDriver[Model]) Filter() common.QueryMod {
	return g.filter
}

func (g GormQueryDriver[Model]) Order() common.QueryMod {
	return g.order
}

func (g GormQueryDriver[Model]) Pagination() common.Pagination {
	return g.pagination
}

func (g GormQueryDriver[Model]) Middleware() []gin.HandlerFunc {
	return g.middleware
}

func (g *GormQueryDriver[Model]) WithFilter(filterFunc GormFilterFunc) *GormQueryDriver[Model] {
	g.filter.modFunc = filterFunc
	return g
}

func (g *GormQueryDriver[Model]) WithPagination(pagination Pagination) *GormQueryDriver[Model] {
	g.pagination.child = pagination
	return g
}

func (g *GormQueryDriver[Model]) WithOrderBy(orderClause any) *GormQueryDriver[Model] {
	g.order.modFunc = func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		return db.Order(orderClause)
	}
	return g
}

func Gorm[Model any](db *gorm.DB) *GormQueryDriver[Model] {
	return &GormQueryDriver[Model]{
		crud: GormQueries[Model](),
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
func GormQueries[Model any]() *crud.CRUD[Model] {
	convertFromDBToInternalValue := FromDBConverter[Model]()

	return &crud.CRUD[Model]{
		List: func(ctx *gin.Context) ([]models.InternalValue, error) {
			rawEntities := []map[string]any{}
			findErr := CtxQuery[Model](ctx).Find(&rawEntities).Error
			if findErr != nil {
				return nil, findErr
			}
			internalValues := make([]models.InternalValue, len(rawEntities))
			for i, rawEntity := range rawEntities {
				internalValue, convertErr := convertFromDBToInternalValue(rawEntity)
				if convertErr != nil {
					return nil, convertErr
				}
				internalValues[i] = internalValue
			}
			return internalValues, findErr
		},
		Retrieve: func(ctx *gin.Context, id any) (models.InternalValue, error) {
			var rawEntity map[string]any
			retrieveErr := CtxQuery[Model](ctx).First(&rawEntity, "id = ?", id).Error
			if retrieveErr != nil {
				if retrieveErr == gorm.ErrRecordNotFound {
					return nil, common.ErrorNotFound
				}
				return nil, retrieveErr
			}
			return convertFromDBToInternalValue(rawEntity)
		},
		Create: func(ctx *gin.Context, m models.InternalValue) (models.InternalValue, error) {
			entity, asModelErr := models.AsModel[Model](m)
			if asModelErr != nil {
				return nil, asModelErr
			}
			createErr := CtxQuery[Model](ctx).Create(&entity).Error
			return models.AsInternalValue(entity), createErr
		},
		Update: func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any) (
			models.InternalValue, error,
		) {
			entity, asModelErr := models.AsModel[Model](new)
			if asModelErr != nil {
				return nil, asModelErr
			}
			updateErr := CtxQuery[Model](ctx).Model(&entity).Updates(&entity).Error
			if updateErr != nil {
				return nil, updateErr
			}
			return models.AsInternalValue(entity), nil
		},
		Delete: func(ctx *gin.Context, id any) error {
			var m Model
			errWrapMsg := "could not delete entity"
			queryResult := CtxQuery[Model](ctx).Delete(&m, "id = ?", id)
			if queryResult.Error != nil {
				return fmt.Errorf(
					"%s: query error: %w", errWrapMsg, queryResult.Error,
				)
			}
			if queryResult.RowsAffected == 0 {
				return common.ErrorNotFound
			}
			return nil
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
