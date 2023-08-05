package views

import (
	"github.com/gin-gonic/gin"

	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/models"
)

func CreateModelFunc[Model any](settings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rawElement map[string]any
		if err := ctx.ShouldBindJSON(&rawElement); err != nil {
			ctx.JSON(400, gin.H{
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
		entity, createErr := settings.Queries.Create(ctx, db.ORM[Model](ctx), &entity)
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
			ctx.JSON(500, gin.H{
				"message": serializeErr.Error(),
			})
			return
		}
		ctx.JSON(201, representation)
	}
}
