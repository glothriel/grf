package serializers

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
)

type SerializerField[Model any] struct {
	fields.Field
	serializer Serializer
}

func (s *SerializerField[Model]) ToRepresentation(iv models.InternalValue, c *gin.Context) (any, error) {
	fieldValue := iv[s.Name()]
	asSlice, isSlice := fieldValue.([]any)
	if isSlice {
		result := make([]any, 0)
		for _, item := range asSlice {
			itemIV, isInternalValue := item.(models.InternalValue)
			if !isInternalValue {
				return nil, fields.NewErrorFieldIsNotPresentInPayload(s.Name())
			}
			serialized, err := s.serializer.ToRepresentation(itemIV, c)
			if err != nil {
				return nil, err
			}
			result = append(result, serialized)
		}
		return result, nil
	}
	return s.serializer.ToRepresentation(iv, c)
}

func (s *SerializerField[Model]) ToInternalValue(raw map[string]any, c *gin.Context) (any, error) {
	return s.serializer.ToInternalValue(raw, c)
}

func NewSerializerField[Model any](name string, serializer Serializer) fields.Field {
	return &SerializerField[Model]{fields.NewField[Model](name), serializer}
}
