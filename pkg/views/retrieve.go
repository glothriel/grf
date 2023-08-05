package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
)

func RetrieveModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		internalValue, retrieveErr := modelSettings.Queries.Retrieve(ctx, db.ORM[Model](ctx), modelSettings.IDFunc(ctx))
		if retrieveErr != nil {
			ctx.JSON(404, gin.H{
				"message": retrieveErr.Error(),
			})
			return
		}
		effectiveSerializer := modelSettings.RetrieveSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}

		formattedElement, toRawErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if toRawErr != nil {
			WriteError(ctx, toRawErr)
			return
		}
		ctx.JSON(200, formattedElement)
	}
}
