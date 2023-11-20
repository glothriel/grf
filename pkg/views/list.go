package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

// ListModelFunc is a gin handler function that lists model instances
func ListModelViewSetFunc[Model any](idf IDFunc, qd queries.Driver[Model], serializer serializers.Serializer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		effectiveSerializer := serializer
		qd.Filter().Apply(ctx)
		qd.Order().Apply(ctx)
		qd.Pagination().Apply(ctx)
		internalValues, listErr := qd.CRUD().List(ctx)
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
		retVal, formatErr := qd.Pagination().Format(ctx, representationItems)
		if formatErr != nil {
			WriteError(ctx, formatErr)
			return
		}
		ctx.JSON(http.StatusOK, retVal)
	}
}
