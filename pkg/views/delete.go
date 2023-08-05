package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
)

func DeleteModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var entity Model
		deleteErr := modelSettings.Queries.Delete(ctx, db.ORM[Model](ctx), entity, modelSettings.IDFunc(ctx))
		if deleteErr != nil {
			ctx.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.JSON(204, nil)
	}
}
