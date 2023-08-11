package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

// ListQueryFunc is a function that executes a list query.
type ListQueryFunc func(ctx *gin.Context) ([]models.InternalValue, error)

// RetrieveQueryFunc is a function that executes a retrieve query.
type RetrieveQueryFunc func(ctx *gin.Context, id any) (models.InternalValue, error)

// CreateQueryFunc is a function that executes a create query.
type CreateQueryFunc func(ctx *gin.Context, new models.InternalValue) (models.InternalValue, error)

// UpdateQueryFunc is a function that executes an update query.
type UpdateQueryFunc func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any) (
	models.InternalValue, error,
)

// DestroyQueryFunc is a function that executes a delete query.
type DestroyQueryFunc func(ctx *gin.Context, id any) error

// CRUD is a collection of queries that are executed by the views
type CRUD[Model any] struct {
	List     ListQueryFunc
	Retrieve RetrieveQueryFunc
	Create   CreateQueryFunc
	Update   UpdateQueryFunc
	Destroy  DestroyQueryFunc
}

func (q *CRUD[Model]) WithList(f ListQueryFunc) *CRUD[Model] {
	q.List = f
	return q
}

func (q *CRUD[Model]) WithRetrieve(f RetrieveQueryFunc) *CRUD[Model] {
	q.Retrieve = f
	return q
}

func (q *CRUD[Model]) WithCreate(f CreateQueryFunc) *CRUD[Model] {
	q.Create = f
	return q
}

func (q *CRUD[Model]) WithUpdate(f UpdateQueryFunc) *CRUD[Model] {
	q.Update = f
	return q
}

func (q *CRUD[Model]) WithDestroy(f DestroyQueryFunc) *CRUD[Model] {
	q.Destroy = f
	return q
}
