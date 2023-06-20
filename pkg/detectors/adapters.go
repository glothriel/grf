package detectors

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/types"
)

func ConvertFuncToRepresentationFuncAdapter(cf types.ConvertFunc) fields.RepresentationFunc {
	return func(intVal models.InternalValue, name string, ctx *gin.Context) (any, error) {
		return cf(intVal[name])
	}
}

func ConvertFuncToInternalValueFuncAdapter(cf types.ConvertFunc) fields.InternalValueFunc {
	return func(reprModel map[string]any, name string, ctx *gin.Context) (any, error) {
		return cf(reprModel[name])
	}
}
