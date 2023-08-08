package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RetrieveModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		modelSettings.QueryDriver.Filter().Apply(ctx)
		internalValue, retrieveErr := modelSettings.QueryDriver.CRUD().Retrieve(ctx, modelSettings.IDFunc(ctx))
		if retrieveErr != nil {
			WriteError(ctx, retrieveErr)
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
		ctx.JSON(http.StatusOK, formattedElement)
	}
}
