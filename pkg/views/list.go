package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/grfctx"
	"gorm.io/gorm"
)

func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context, dbSession *gorm.DB) {

		var entities []map[string]any
		modelSettings.Pagination.Apply(
			ctx.Gin,
			modelSettings.OrderBy(ctx.Gin, modelSettings.Filter(ctx.Gin, dbSession)),
		).Find(&entities)
		representationItems := []any{}
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		for _, entity := range entities {
			internalValue, internalValueErr := effectiveSerializer.FromDB(entity, ctx)
			if internalValueErr != nil {
				WriteError(ctx.Gin, internalValueErr)
				return
			}
			rawElement, toRawErr := effectiveSerializer.ToRepresentation(
				internalValue, ctx,
			)

			if toRawErr != nil {
				ctx.Gin.JSON(500, gin.H{
					"message": toRawErr.Error(),
				})
				return
			}
			representationItems = append(representationItems, rawElement)
		}
		ctx.Gin.JSON(200, representationItems)
	}
}
