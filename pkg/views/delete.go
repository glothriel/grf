package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
)

func DeleteModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var entity Model
		deleteErr := db.ORM[Model](ctx).Delete(&entity, "id = ?", modelSettings.IDFunc(ctx)).Error
		if deleteErr != nil {
			ctx.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.JSON(204, nil)
	}
}
