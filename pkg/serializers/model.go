package serializers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/glothriel/gin-rest-framework/pkg/fields"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/types"
	"github.com/sirupsen/logrus"
)

type ModelSerializer[Model any] struct {
	Fields          map[string]*fields.Field[Model]
	FieldTypeMapper *types.FieldTypeMapper
}

func (s *ModelSerializer[Model]) ToInternalValue(raw map[string]interface{}) (*models.InternalValue[Model], error) {
	intVMap := make(map[string]interface{})
	superfluousFields := make([]string, 0)
	for k := range raw {
		field, ok := s.Fields[k]
		if !ok {
			superfluousFields = append(superfluousFields, k)
			continue
		}
		if !field.Writable {
			continue
		}
		intV, err := field.InternalValueFunc(raw, k)
		if err != nil {
			return nil, &ValidationError{FieldErrors: map[string][]string{k: {err.Error()}}}
		}
		intVMap[k] = intV
	}
	if len(superfluousFields) > 0 {
		errMap := map[string][]string{}
		existingFields := []string{}
		for _, field := range s.Fields {
			existingFields = append(existingFields, field.Name())
		}
		for _, field := range superfluousFields {
			errMap[field] = []string{fmt.Sprintf("Field `%s` is not accepted by this endpoint, accepted fields: %s", field, strings.Join(existingFields, ", "))}
		}
		return nil, &ValidationError{FieldErrors: errMap}
	}
	return &models.InternalValue[Model]{Map: intVMap}, nil
}

func (s *ModelSerializer[Model]) ToRepresentation(intVal *models.InternalValue[Model]) (map[string]interface{}, error) {
	raw := make(map[string]interface{})
	for _, field := range s.Fields {
		if !field.Readable {
			continue
		}
		value, err := field.ToRepresentation(intVal)

		if err != nil {
			return nil, fmt.Errorf(
				"Failed to serialize field `%s` to representation: %w", field.Name(), err,
			)
		}
		raw[field.Name()] = value
	}
	return raw, nil
}

func (s *ModelSerializer[Model]) FromDB(raw map[string]interface{}) (*models.InternalValue[Model], error) {

	intVMap := make(map[string]interface{})
	for k := range raw {
		field, ok := s.Fields[k]
		if !ok {
			continue
		}
		intV, err := field.FromDBFunc(raw, k)
		if err != nil {
			return nil, fmt.Errorf("Failed to serialize field `%s` from database: %s", k, err)
		}
		intVMap[k] = intV
	}

	return &models.InternalValue[Model]{Map: intVMap}, nil
}

func (s *ModelSerializer[Model]) Validate(intVal *models.InternalValue[Model]) error {
	return nil
}

func (s *ModelSerializer[Model]) WithField(field *fields.Field[Model]) *ModelSerializer[Model] {
	s.Fields[field.Name()] = field
	return s
}

func (s *ModelSerializer[Model]) WithFieldUpdated(name string, updateFunc func(f *fields.Field[Model])) *ModelSerializer[Model] {
	v, ok := s.Fields[name]
	if !ok {
		var m Model
		logrus.Fatalf("Could not find field `%s` on model `%s` when registering serializer field", name, reflect.TypeOf(m))
	}
	updateFunc(v)
	return s

}

func (s *ModelSerializer[Model]) WithExistingFields(passedFields []string) *ModelSerializer[Model] {
	s.Fields = make(map[string]*fields.Field[Model])
	var m Model
	attributeTypes := DetectAttributes(m)
	for _, field := range passedFields {
		attributeType, ok := attributeTypes[field]
		if !ok {
			logrus.Fatalf("Could not find field `%s` on model `%s` when registering serializer field", field, reflect.TypeOf(m))
		}
		toRepresentation, toRepresentationErr := s.FieldTypeMapper.ToRepresentation(attributeType)
		if toRepresentationErr != nil {
			logrus.Fatalf("Could not determine representation of field `%s` on model `%s`: %s", field, reflect.TypeOf(m), toRepresentationErr)
		}
		toInternalValue, toInternalValueErr := s.FieldTypeMapper.ToInternalValue(attributeType)
		if toInternalValueErr != nil {
			logrus.Fatalf("Could not determine internal value of field `%s` on model `%s`: %s", field, reflect.TypeOf(m), toInternalValueErr)
		}
		s.Fields[field] = fields.NewField[Model](
			field,
		).WithRepresentationFunc(
			ConvertFuncToRepresentationFuncAdapter[Model](toRepresentation),
		).WithInternalValueFunc(
			ConvertFuncToInternalValueFuncAdapter(toInternalValue),
		)
	}
	return s
}

func NewModelSerializer[Model any](ftm *types.FieldTypeMapper) *ModelSerializer[Model] {
	fieldNames := []string{}
	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous {
			fieldNames = append(fieldNames, field.Tag.Get("json"))
		}
	}

	return (&ModelSerializer[Model]{
		FieldTypeMapper: ftm,
	}).WithExistingFields(fieldNames)
}

func ConvertFuncToRepresentationFuncAdapter[Model any](cf types.ConvertFunc) fields.RepresentationFunc[Model] {
	return func(intVal *models.InternalValue[Model], name string) (interface{}, error) {
		return cf(intVal.Map[name])
	}
}

func ConvertFuncToInternalValueFuncAdapter(cf types.ConvertFunc) fields.InternalValueFunc {
	return func(reprModel map[string]interface{}, name string) (interface{}, error) {
		return cf(reprModel[name])
	}
}

// Prints a summary with the fields of the model obtained using reflection
func DetectAttributes[Model any](model Model) map[string]string {
	ret := make(map[string]string)
	fields := reflect.VisibleFields(reflect.TypeOf(model))
	for _, field := range fields {
		if !field.Anonymous {
			ret[field.Tag.Get("json")] = field.Type.String()
		}
	}
	return ret
}
