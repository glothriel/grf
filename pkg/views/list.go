package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
)

func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var entities []map[string]any
		modelSettings.Pagination.Apply(
			ctx,
			modelSettings.OrderBy(ctx, modelSettings.Filter(ctx, db.ORM[Model](ctx))),
		).Find(&entities)
		representationItems := []any{}
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		for _, entity := range entities {
			internalValue, internalValueErr := effectiveSerializer.FromDB(entity, ctx)
			if internalValueErr != nil {
				WriteError(ctx, internalValueErr)
				return
			}
			rawElement, toRawErr := effectiveSerializer.ToRepresentation(
				internalValue, ctx,
			)

			if toRawErr != nil {
				ctx.JSON(500, gin.H{
					"message": toRawErr.Error(),
				})
				return
			}
			representationItems = append(representationItems, rawElement)
		}
		ctx.JSON(200, representationItems)
	}
}
