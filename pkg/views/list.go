package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListModelFunc is a gin handler function that lists model instances
func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		modelSettings.QueryDriver.Filter().Apply(ctx)
		modelSettings.QueryDriver.Order().Apply(ctx)
		modelSettings.QueryDriver.Pagination().Apply(ctx)
		internalValues, listErr := modelSettings.QueryDriver.CRUD().List(ctx)
		if listErr != nil {
			WriteError(ctx, listErr)
			return
		}
		representationItems := []any{}
		for _, internalValue := range internalValues {
			rawElement, toRawErr := effectiveSerializer.ToRepresentation(
				internalValue, ctx,
			)
			if toRawErr != nil {
				WriteError(ctx, toRawErr)
				return
			}
			representationItems = append(representationItems, rawElement)
		}
		retVal, formatErr := modelSettings.QueryDriver.Pagination().Format(ctx, representationItems)
		if formatErr != nil {
			WriteError(ctx, formatErr)
			return
		}
		ctx.JSON(http.StatusOK, retVal)
	}
}
