package serializers

import "github.com/glothriel/gin-rest-framework/pkg/models"

type Serializer[Model any] interface {
	ToInternalValue(map[string]interface{}) (*models.InternalValue[Model], error)
	FromDB(map[string]interface{}) (*models.InternalValue[Model], error)
	Validate(*models.InternalValue[Model]) error
	ToRepresentation(*models.InternalValue[Model]) (map[string]interface{}, error)
}
