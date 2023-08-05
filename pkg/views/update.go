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
		updates["id"] = modelSettings.IDFunc(ctx)
		effectiveSerializer := modelSettings.UpdateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		incomingIntVal, fromRawErr := effectiveSerializer.ToInternalValue(updates, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}

		oldIntVal, oldErr := modelSettings.Queries.Retrieve(ctx, db.ORM[Model](ctx), modelSettings.IDFunc(ctx))
		if oldErr != nil {
			ctx.JSON(404, gin.H{
				"message": oldErr.Error(),
			})
			return
		}
		oldEntity, asModelErr := models.AsModel[Model](oldIntVal)
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		for k, v := range incomingIntVal {
			oldIntVal[k] = v
		}
		entity, asModelErr := models.AsModel[Model](oldIntVal)
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		updated, updateErr := modelSettings.Queries.Update(ctx, db.ORM[Model](ctx), &oldEntity, &entity, modelSettings.IDFunc(ctx))
		if updateErr != nil {
			WriteError(ctx, updateErr)
			return
		}
		internalValue, intValErr := models.AsInternalValue(updated)
		if intValErr != nil {
			WriteError(ctx, intValErr)
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
