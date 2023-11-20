package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

func RetrieveModelViewSetFunc[Model any](idf IDFunc, qd queries.Driver[Model], serializer serializers.Serializer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		qd.Filter().Apply(ctx)
		internalValue, retrieveErr := qd.CRUD().Retrieve(ctx, idf(ctx))
		if retrieveErr != nil {
			WriteError(ctx, retrieveErr)
			return
		}
		effectiveSerializer := serializer

		formattedElement, toRawErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if toRawErr != nil {
			WriteError(ctx, toRawErr)
			return
		}
		ctx.JSON(http.StatusOK, formattedElement)
	}
}
