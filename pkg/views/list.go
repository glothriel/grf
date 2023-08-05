package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
)

func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		internalValues, _, listErr := modelSettings.Queries.List(ctx, modelSettings.Pagination.Apply(
			ctx,
			modelSettings.OrderBy(ctx, modelSettings.Filter(ctx, db.ORM[Model](ctx))),
		))
		if listErr != nil {
			ctx.JSON(500, gin.H{
				"message": listErr.Error(),
			})
			return
		}
		representationItems := []map[string]any{}
		for _, internalValue := range internalValues {
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
