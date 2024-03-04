// Package views provides a set of functions that can be used to create views for models.
package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

// CreateModelFunc is a function that creates a new model
func CreateModelViewSetFunc[Model any](idf IDFunc, qd queries.Driver[Model], serializer serializers.Serializer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rawElement map[string]any
		if parseErr := ctx.ShouldBindJSON(&rawElement); parseErr != nil {
			WriteError(ctx, parseErr)
			return
		}
		internalValue, fromRawErr := serializer.ToInternalValue(rawElement, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		internalValue, createErr := qd.CRUD().Create(ctx, internalValue)
		if createErr != nil {
			WriteError(ctx, createErr)
			return
		}
		representation, serializeErr := serializer.ToRepresentation(internalValue, ctx)
		if serializeErr != nil {
			WriteError(ctx, serializeErr)
			return
		}
		ctx.JSON(http.StatusCreated, representation)
	}
}
