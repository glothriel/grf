package serializers

import "github.com/mitchellh/mapstructure"

// uses mapstructure to convert map[string]interface{} to Model
type PassThroughSerializer[Model any] struct{}

func (s *PassThroughSerializer[Model]) ToInternalValue(raw interface{}) (Model, error) {
	var entity Model
	return entity, mapstructure.Decode(raw, &entity)
}

func (s *PassThroughSerializer[Model]) ToRepresentation(entity Model) (interface{}, error) {
	return entity, nil
}

func (s *PassThroughSerializer[Model]) Validate(entity Model) error {
	return nil
}
