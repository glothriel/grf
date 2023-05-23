package serializers

import "github.com/glothriel/gin-rest-framework/pkg/models"

type Serializer[Model any] interface {
	ToInternalValue(map[string]any) (models.InternalValue[Model], error)
	FromDB(map[string]any) (models.InternalValue[Model], error)
	Validate(models.InternalValue[Model]) error
	ToRepresentation(models.InternalValue[Model]) (map[string]any, error)
}
