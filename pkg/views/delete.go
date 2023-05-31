package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/grfctx"
	"gorm.io/gorm"
)

func DeleteModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context, dbSession *gorm.DB) {

		var entity Model
		deleteErr := dbSession.Delete(&entity, "id = ?", modelSettings.IDFunc(ctx.Gin)).Error
		if deleteErr != nil {
			ctx.Gin.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.Gin.JSON(204, nil)
	}
}
