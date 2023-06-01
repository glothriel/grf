package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/models"
)

func UpdateModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var updates map[string]any
		if err := ctx.ShouldBindJSON(&updates); err != nil {
			ctx.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		updates["id"] = ctx.Param("id")
		effectiveSerializer := modelSettings.UpdateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		intVal, fromRawErr := effectiveSerializer.ToInternalValue(updates, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		entity, asModelErr := models.AsModel[Model](intVal)
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		updateErr := db.ORM[Model](ctx).Model(&entity).Updates(&entity).Error
		if updateErr != nil {
			WriteError(ctx, updateErr)
			return
		}
		var updatedMap map[string]any
		if err := db.ORM[Model](ctx).First(&updatedMap, "id = ?", modelSettings.IDFunc(ctx)).Error; err != nil {
			ctx.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		internalValue, internalValueErr := effectiveSerializer.FromDB(updatedMap, ctx)
		if internalValueErr != nil {
			WriteError(ctx, internalValueErr)
			return
		}
		rawElement, toRawErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if toRawErr != nil {
			ctx.JSON(500, gin.H{
				"message": toRawErr.Error(),
			})
			return
		}
		ctx.JSON(200, rawElement)
	}
}
