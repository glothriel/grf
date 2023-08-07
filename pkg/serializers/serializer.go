package serializers

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

type Representation map[string]any

type Serializer interface {
	ToInternalValue(map[string]any, *gin.Context) (models.InternalValue, error)
	ToRepresentation(models.InternalValue, *gin.Context) (Representation, error)
}
