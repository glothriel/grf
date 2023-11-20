package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
)

func DestroyModelViewSetFunc[Model any](idf IDFunc, qd queries.Driver[Model], serializer serializers.Serializer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		deleteErr := qd.CRUD().Destroy(ctx, idf(ctx))
		if deleteErr != nil {
			WriteError(ctx, deleteErr)
			return
		}
		ctx.JSON(http.StatusNoContent, nil)
	}
}
