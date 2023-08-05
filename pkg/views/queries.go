package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"gorm.io/gorm"
)

// ListQueryFunc is a function that executes a list query.
type ListQueryFunc[Model any] func(ctx *gin.Context, db *gorm.DB) ([]models.InternalValue, ListQueryContext, error)

// RetrieveQueryFunc is a function that executes a retrieve query.
type RetrieveQueryFunc[Model any] func(ctx *gin.Context, db *gorm.DB, id any) (models.InternalValue, error)

// CreateQueryFunc is a function that executes a create query.
type CreateQueryFunc[Model any] func(ctx *gin.Context, db *gorm.DB, m *Model) (Model, error)

// UpdateQueryFunc is a function that executes an update query.
type UpdateQueryFunc[Model any] func(ctx *gin.Context, db *gorm.DB, old *Model, new *Model, id any) (
	Model, error,
)

// DeleteQueryFunc is a function that executes a delete query.
type DeleteQueryFunc[Model any] func(ctx *gin.Context, db *gorm.DB, m Model, id any) error

// ListQueryContext can be used to pass some information from the query to pagination implementations
type ListQueryContext map[string]any

// Queries is a collection of queries that are executed by the views
type Queries[Model any] struct {
	List     ListQueryFunc[Model]
	Retrieve RetrieveQueryFunc[Model]
	Create   CreateQueryFunc[Model]
	Update   UpdateQueryFunc[Model]
	Delete   DeleteQueryFunc[Model]
}

// DefaultQueries returns default queries providing basic CRUD functionality
func DefaultQueries[Model any]() Queries[Model] {
	converter := FromDBConverter[Model]()
	var emptyModelForDBModel Model

	return Queries[Model]{
		List: func(ctx *gin.Context, db *gorm.DB) ([]models.InternalValue, ListQueryContext, error) {
			rawEntities := []map[string]any{}
			findErr := db.Model(&emptyModelForDBModel).Find(&rawEntities).Error
			if findErr != nil {
				return nil, ListQueryContext{}, findErr
			}
			internalValues := make([]models.InternalValue, len(rawEntities))
			for i, rawEntity := range rawEntities {
				internalValue, convertErr := converter(rawEntity)
				if convertErr != nil {
					return nil, ListQueryContext{}, convertErr
				}
				internalValues[i] = internalValue
			}
			return internalValues, ListQueryContext{}, findErr
		},
		Retrieve: func(ctx *gin.Context, db *gorm.DB, id any) (models.InternalValue, error) {
			var rawEntity map[string]any
			retrieveErr := db.Model(&emptyModelForDBModel).First(&rawEntity, "id = ?", id).Error
			if retrieveErr != nil {
				return nil, retrieveErr
			}
			return converter(rawEntity)
		},
		Create: func(ctx *gin.Context, db *gorm.DB, m *Model) (Model, error) {
			createErr := db.Create(m).Error
			return *m, createErr
		},
		Update: func(ctx *gin.Context, db *gorm.DB, old *Model, new *Model, id any) (Model, error) {
			updateErr := db.Model(new).Updates(new).Error
			if updateErr != nil {
				return *new, updateErr
			}
			return *new, nil
		},
		Delete: func(ctx *gin.Context, db *gorm.DB, m Model, id any) error {
			deleteErr := db.Delete(&m, "id = ?", id).Error
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
