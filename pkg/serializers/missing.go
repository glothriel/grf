package serializers

import (
	"fmt"
	"reflect"

	"github.com/glothriel/gin-rest-framework/pkg/models"
)

type MissingSerializer[Model any] struct {
}

func (s *MissingSerializer[Model]) ToInternalValue(raw map[string]any) (*models.InternalValue[Model], error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}

func (s *MissingSerializer[Model]) FromDB(raw map[string]any) (*models.InternalValue[Model], error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}

func (s *MissingSerializer[Model]) ToRepresentation(intVal *models.InternalValue[Model]) (map[string]any, error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}

func (s *MissingSerializer[Model]) Validate(intVal *models.InternalValue[Model]) error {
	var m Model
	return fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}
