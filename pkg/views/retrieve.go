package views

import (
	"github.com/gin-gonic/gin"
)

func RetrieveModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		modelSettings.Database.Filter().Apply(ctx)
		internalValue, retrieveErr := modelSettings.Database.Queries().Retrieve(ctx, modelSettings.IDFunc(ctx))
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
