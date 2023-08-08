package common

import "github.com/gin-gonic/gin"

type QueryMod interface {
	Apply(*gin.Context)
}

type Pagination interface {
	QueryMod
	Format(*gin.Context, []any) (any, error)
}
