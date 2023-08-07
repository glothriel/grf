package db

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

// ListQueryFunc is a function that executes a list query.
type ListQueryFunc[Model any] func(ctx *gin.Context) ([]models.InternalValue, error)

// RetrieveQueryFunc is a function that executes a retrieve query.
type RetrieveQueryFunc[Model any] func(ctx *gin.Context, id any) (models.InternalValue, error)

// CreateQueryFunc is a function that executes a create query.
type CreateQueryFunc[Model any] func(ctx *gin.Context, m *Model) (Model, error)

// UpdateQueryFunc is a function that executes an update query.
type UpdateQueryFunc[Model any] func(ctx *gin.Context, old *Model, new *Model, id any) (
	Model, error,
)

// DeleteQueryFunc is a function that executes a delete query.
type DeleteQueryFunc[Model any] func(ctx *gin.Context, m Model, id any) error

// Queries is a collection of queries that are executed by the views
type Queries[Model any] struct {
	List     ListQueryFunc[Model]
	Retrieve RetrieveQueryFunc[Model]
	Create   CreateQueryFunc[Model]
	Update   UpdateQueryFunc[Model]
	Delete   DeleteQueryFunc[Model]
}

func (q *Queries[Model]) WithList(f ListQueryFunc[Model]) *Queries[Model] {
	q.List = f
	return q
}

func (q *Queries[Model]) WithRetrieve(f RetrieveQueryFunc[Model]) *Queries[Model] {
	q.Retrieve = f
	return q
}

func (q *Queries[Model]) WithCreate(f CreateQueryFunc[Model]) *Queries[Model] {
	q.Create = f
	return q
}

func (q *Queries[Model]) WithUpdate(f UpdateQueryFunc[Model]) *Queries[Model] {
	q.Update = f
	return q
}

func (q *Queries[Model]) WithDelete(f DeleteQueryFunc[Model]) *Queries[Model] {
	q.Delete = f
	return q
}
