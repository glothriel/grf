package queries

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/queries/crud"
)

type Driver[Model any] interface {
	CRUD() *crud.CRUD[Model]

	Pagination() common.Pagination
	Filter() common.QueryMod
	Order() common.QueryMod

	Middleware() []gin.HandlerFunc
}
