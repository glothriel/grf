package views

import (
	"github.com/gin-gonic/gin"
)

func DeleteModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var entity Model
		deleteErr := modelSettings.Database.Queries().Delete(ctx, entity, modelSettings.IDFunc(ctx))
		if deleteErr != nil {
			ctx.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.JSON(204, nil)
	}
}
