package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/grfctx"
)

func DeleteModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context) {

		var entity Model
		deleteErr := db.CtxNewQuery[Model](ctx).Delete(&entity, "id = ?", modelSettings.IDFunc(ctx)).Error
		if deleteErr != nil {
			ctx.Gin.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.Gin.JSON(204, nil)
	}
}
