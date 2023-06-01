package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/glothriel/grf/pkg/models"
)

func UpdateModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context) {

		var updates map[string]any
		if err := ctx.Gin.ShouldBindJSON(&updates); err != nil {
			ctx.Gin.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		updates["id"] = ctx.Gin.Param("id")
		effectiveSerializer := modelSettings.UpdateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		intVal, fromRawErr := effectiveSerializer.ToInternalValue(updates, ctx)
		if fromRawErr != nil {
			WriteError(ctx.Gin, fromRawErr)
			return
		}
		entity, asModelErr := models.AsModel[Model](intVal)
		if asModelErr != nil {
			WriteError(ctx.Gin, asModelErr)
			return
		}
		updateErr := db.CtxNewQuery[Model](ctx).Model(&entity).Updates(&entity).Error
		if updateErr != nil {
			WriteError(ctx.Gin, updateErr)
			return
		}
		var updatedMap map[string]any
		if err := db.CtxNewQuery[Model](ctx).First(&updatedMap, "id = ?", modelSettings.IDFunc(ctx)).Error; err != nil {
			ctx.Gin.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		internalValue, internalValueErr := effectiveSerializer.FromDB(updatedMap, ctx)
		if internalValueErr != nil {
			WriteError(ctx.Gin, internalValueErr)
			return
		}
		rawElement, toRawErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if toRawErr != nil {
			ctx.Gin.JSON(500, gin.H{
				"message": toRawErr.Error(),
			})
			return
		}
		ctx.Gin.JSON(200, rawElement)
	}
}
