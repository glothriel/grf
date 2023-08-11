package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DestroyModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		deleteErr := modelSettings.QueryDriver.CRUD().Destroy(ctx, modelSettings.IDFunc(ctx))
		if deleteErr != nil {
			WriteError(ctx, deleteErr)
			return
		}
		ctx.JSON(http.StatusNoContent, nil)
	}
}
