package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/glothriel/grf/pkg/models"
	"gorm.io/gorm"
)

func RetrieveModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context, dbSession *gorm.DB) {

		var entity Model
		if err := modelSettings.Filter(ctx.Gin, dbSession.First(&entity, "id = ?", modelSettings.IDFunc(ctx.Gin))).Error; err != nil {
			ctx.Gin.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		effectiveSerializer := modelSettings.RetrieveSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}

		internalValue, internalValueErr := models.AsInternalValue(entity)
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
