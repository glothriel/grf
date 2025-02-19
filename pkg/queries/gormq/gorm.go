package gormq

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/detectors"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/queries/crud"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GormFilterFunc func(ctx *gin.Context, db *gorm.DB) *gorm.DB

type gormQueryMod[Model any] struct {
	modFunc GormFilterFunc
}

func (g gormQueryMod[Model]) Apply(ctx *gin.Context) {
	CtxSetQuery(ctx, g.modFunc(ctx, CtxQuery(ctx)))
}

type gormPagination[Model any] struct {
	child Pagination
}

func (g gormPagination[Model]) Apply(ctx *gin.Context) {
	CtxSetQuery(ctx, g.child.Apply(ctx, CtxQuery(ctx)))
}

func (g gormPagination[Model]) Format(ctx *gin.Context, elems []any) (any, error) {
	return g.child.Format(ctx, elems)
}

type GormQueryDriver[Model any] struct {
	filter           *gormQueryMod[Model]
	preloads         *gormQueryMod[Model]
	fieldNames       map[string]string
	preloadedQueries []string
	order            *gormQueryMod[Model]
	pagination       *gormPagination[Model]

	middleware []gin.HandlerFunc
}

func (g GormQueryDriver[Model]) CRUD() *crud.CRUD[Model] {
	return GormQueries[Model](g.preloadedQueries)
}

func (g GormQueryDriver[Model]) Filter() common.QueryMod {
	return common.NewCompositeQueryMod(g.filter, g.preloads)
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

func (g *GormQueryDriver[Model]) WithPreload(query string, args ...any) *GormQueryDriver[Model] {
	g.preloads.modFunc = func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		fieldName, ok := g.fieldNames[query]
		if !ok {
			logrus.Errorf("Could not find field name for query %s, skipping preload", query)
		} else {
			db = db.Preload(fieldName, args...)
		}
		return db
	}
	g.preloadedQueries = append(g.preloadedQueries, query)
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

func Gorm[Model any](factory GormORMFactory) *GormQueryDriver[Model] {
	return &GormQueryDriver[Model]{
		preloadedQueries: []string{},
		fieldNames:       detectors.FieldNames[Model](),
		filter: &gormQueryMod[Model]{
			modFunc: func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				return db
			},
		},
		preloads: &gormQueryMod[Model]{
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
				CtxSetFactory(ctx, factory)
				CtxInitQuery(ctx)
				ctx.Next()
			},
		},
	}
}

// GormQueries returns default queries providing basic CRUD functionality
func GormQueries[Model any](preloadedQueries []string) *crud.CRUD[Model] {
	ConvertFromDBToInternalValue := FromDBConverter[Model]()
	var empty Model
	var preloadedQueriesMap = make(map[string]bool)
	for _, query := range preloadedQueries {
		preloadedQueriesMap[query] = true
	}
	return &crud.CRUD[Model]{
		List: func(ctx *gin.Context) ([]models.InternalValue, error) {
			rawEntities := []models.InternalValue{}
			typedEntities := []Model{}
			findErr := CtxQuery(ctx).Model(&empty).Find(&typedEntities).Error
			if findErr != nil {
				return nil, findErr
			}
			for _, entity := range typedEntities {
				iv := models.AsInternalValue(entity)
				for k, v := range iv {
					if _, ok := preloadedQueriesMap[k]; ok {
						vValue := reflect.ValueOf(v)

						if vValue.Kind() == reflect.Slice {
							newSlice := make([]any, vValue.Len())

							for i := 0; i < vValue.Len(); i++ {
								newSlice[i] = models.AsInternalValue(vValue.Index(i).Interface())
							}

							iv[k] = newSlice
						} else {
							iv[k] = models.AsInternalValue(v)
						}
					} else {
						iv[k] = v
					}
				}
				rawEntities = append(rawEntities, iv)
			}
			return rawEntities, findErr
		},
		Retrieve: func(ctx *gin.Context, id any) (models.InternalValue, error) {
			var rawEntity map[string]any
			retrieveErr := CtxQuery(ctx).Model(&empty).First(&rawEntity, "id = ?", id).Error
			if retrieveErr != nil {
				if retrieveErr == gorm.ErrRecordNotFound {
					return nil, common.ErrorNotFound
				}
				return nil, retrieveErr
			}
			return ConvertFromDBToInternalValue(rawEntity)
		},
		Create: func(ctx *gin.Context, m models.InternalValue) (models.InternalValue, error) {
			entity, asModelErr := models.AsModel[Model](m)
			if asModelErr != nil {
				return nil, asModelErr
			}
			createErr := CtxQuery(ctx).Model(&empty).Create(&entity).Error
			return models.AsInternalValue(entity), createErr
		},
		Update: func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any) (
			models.InternalValue, error,
		) {
			entity, asModelErr := models.AsModel[Model](new)
			if asModelErr != nil {
				return nil, asModelErr
			}
			updateErr := CtxQuery(ctx).Model(&entity).Updates(&entity).Error
			if updateErr != nil {
				return nil, updateErr
			}
			return models.AsInternalValue(entity), nil
		},
		Destroy: func(ctx *gin.Context, id any) error {
			var m Model
			errWrapMsg := "could not delete entity"
			queryResult := CtxQuery(ctx).Model(&empty).Delete(&m, "id = ?", id)
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
// as GORM does this only for structs
func FromDBConverter[Model any]() func(map[string]any) (models.InternalValue, error) {
	fromDB := SQLScannerOrPassthrough[Model]()
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
