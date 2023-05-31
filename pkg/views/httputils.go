package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/sirupsen/logrus"
)

func WriteError(ctx *gin.Context, err error) {
	ve, isValidationErr := err.(*serializers.ValidationError)
	if isValidationErr {
		ctx.JSON(400, gin.H{
			"errors": ve.FieldErrors,
		})
		return
	}
	logrus.Error(err)
	ctx.JSON(500, "Internal Server Error")
}
