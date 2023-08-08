package views

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/sirupsen/logrus"
)

// WriteError checks for common error types and maps them to correct HTTP status codes
func WriteError(ctx *gin.Context, err error) {
	// Serializers validation
	ve, isValidationErr := err.(*serializers.ValidationError)
	if isValidationErr {
		ctx.JSON(400, gin.H{
			"errors": ve.FieldErrors,
		})
		return
	}
	// QueryDriver returns common.ErrorNotFound when no entity is found
	if errors.Is(err, common.ErrorNotFound) {
		ctx.JSON(404, gin.H{
			"message": err.Error(),
		})
		return
	}
	// Empty JSON body or JSON syntax error
	_, isSyntaxErr := err.(*json.SyntaxError)
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) || isSyntaxErr {
		ctx.JSON(400, gin.H{
			"errors": map[string][]string{
				"all": {"could not parse request body"},
			},
		})
		return
	}
	logrus.Errorf("Unexpected error of type %T: %s", err, err.Error())
	ctx.JSON(500, gin.H{
		"message": "internal server error",
	})
}
