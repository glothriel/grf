package views

import (
	"github.com/gin-gonic/gin"
)

func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		modelSettings.Database.Filter().Apply(ctx)
		modelSettings.Database.Order().Apply(ctx)
		modelSettings.Database.Pagination().Apply(ctx)
		internalValues, listErr := modelSettings.Database.Queries().List(ctx)
		if listErr != nil {
			ctx.JSON(500, gin.H{
				"message": listErr.Error(),
			})
			return
		}
		representationItems := []any{}
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
		retVal, formatErr := modelSettings.Database.Pagination().Format(ctx, representationItems)
		if formatErr != nil {
			ctx.JSON(500, gin.H{
				"message": formatErr.Error(),
			})
			return
		}
		ctx.JSON(200, retVal)
	}
}
