package views

import (
	"github.com/gin-gonic/gin"

	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

type IDFunc func(*gin.Context) string

type ViewsetHandlerFactoryFunc[Model any] func(IDFunc, queries.Driver[Model], serializers.Serializer) gin.HandlerFunc

func IDFromQueryParamIDFunc(ctx *gin.Context) string {
	return ctx.Param("id")
}
