package views

import (
	"github.com/gin-gonic/gin"

	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

type IDFunc func(*gin.Context) string

type ViewSetHandlerFunc[Model any] func(IDFunc, queries.Driver[Model], serializers.Serializer) gin.HandlerFunc

type ViewSetHandlerFactoryFunc[Model any] ViewSetHandlerFunc[Model]

func IDFromQueryParamIDFunc(ctx *gin.Context) string {
	return ctx.Param("id")
}

func IDFromPathParam(paramName string) func(ctx *gin.Context) string {
	return func(ctx *gin.Context) string {
		return ctx.Param(paramName)
	}
}
