package db

import (
	"github.com/gin-gonic/gin"
)

type Pagination interface {
	Apply(*gin.Context)
	Format(*gin.Context, []any) (any, error)
}

type QueryMod interface {
	Apply(*gin.Context)
}

type Database[Model any] interface {
	Pagination() Pagination
	Filter() QueryMod
	Order() QueryMod

	Queries() *Queries[Model]

	Middleware() []gin.HandlerFunc
}
