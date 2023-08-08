package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

// ListQueryFunc is a function that executes a list query.
type ListQueryFunc[Model any] func(ctx *gin.Context) ([]models.InternalValue, error)

// RetrieveQueryFunc is a function that executes a retrieve query.
type RetrieveQueryFunc[Model any] func(ctx *gin.Context, id any) (models.InternalValue, error)

// CreateQueryFunc is a function that executes a create query.
type CreateQueryFunc[Model any] func(ctx *gin.Context, new models.InternalValue) (models.InternalValue, error)

// UpdateQueryFunc is a function that executes an update query.
type UpdateQueryFunc[Model any] func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any) (
	models.InternalValue, error,
)

// DeleteQueryFunc is a function that executes a delete query.
type DeleteQueryFunc[Model any] func(ctx *gin.Context, id any) error

// CRUD is a collection of queries that are executed by the views
type CRUD[Model any] struct {
	List     ListQueryFunc[Model]
	Retrieve RetrieveQueryFunc[Model]
	Create   CreateQueryFunc[Model]
	Update   UpdateQueryFunc[Model]
	Delete   DeleteQueryFunc[Model]
}

func (q *CRUD[Model]) WithList(f ListQueryFunc[Model]) *CRUD[Model] {
	q.List = f
	return q
}

func (q *CRUD[Model]) WithRetrieve(f RetrieveQueryFunc[Model]) *CRUD[Model] {
	q.Retrieve = f
	return q
}

func (q *CRUD[Model]) WithCreate(f CreateQueryFunc[Model]) *CRUD[Model] {
	q.Create = f
	return q
}

func (q *CRUD[Model]) WithUpdate(f UpdateQueryFunc[Model]) *CRUD[Model] {
	q.Update = f
	return q
}

func (q *CRUD[Model]) WithDelete(f DeleteQueryFunc[Model]) *CRUD[Model] {
	q.Delete = f
	return q
}
