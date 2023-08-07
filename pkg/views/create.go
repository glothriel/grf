// Package views provides a set of functions that can be used to create views for models.
package views

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/glothriel/grf/pkg/models"
)

// CreateModelFunc is a function that creates a new model
func CreateModelFunc[Model any](settings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rawElement map[string]any
		if err := ctx.ShouldBindJSON(&rawElement); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		effectiveSerializer := settings.CreateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = settings.DefaultSerializer
		}
		internalValue, fromRawErr := effectiveSerializer.ToInternalValue(rawElement, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		// Gorm supports creating rows using maps, but we cannot use that, because in that case
		// Gorm won't execute hooks. UUID-based PKs require a hook to be executed. That's why we
		// convert the map to a struct and execute the query, despite reflection being slow.
		entity, asModelErr := models.AsModel[Model](internalValue)
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		entity, createErr := settings.Database.Queries().Create(ctx, &entity)
		if createErr != nil {
			WriteError(ctx, createErr)
			return
		}
		internalValue, internalValueErr := models.AsInternalValue(entity)
		if internalValueErr != nil {
			WriteError(ctx, internalValueErr)
			return
		}
		representation, serializeErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if serializeErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": serializeErr.Error(),
			})
			return
		}
		ctx.JSON(http.StatusCreated, representation)
	}
}
