package serializers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/types"
	"github.com/sirupsen/logrus"
)

type ModelSerializer[Model any] struct {
	Fields          map[string]Field[Model]
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
		value, err := field.ToRepresentation(intVal)

		if err != nil {
			return nil, err
		}
		raw[field.Name()] = value
	}
	return raw, nil
}

func (s *ModelSerializer[Model]) Validate(intVal *models.InternalValue[Model]) error {
	return nil
}

func (s *ModelSerializer[Model]) WithField(field Field[Model]) *ModelSerializer[Model] {
	s.Fields[field.Name()] = field
	return s
}

func (s *ModelSerializer[Model]) WithExistingFields(fields []string) *ModelSerializer[Model] {
	s.Fields = make(map[string]Field[Model])
	var m Model
	attributeTypes := DetectAttributes(m)
	for _, field := range fields {
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
		s.Fields[field] = Field[Model]{
			ItsName:            field,
			RepresentationFunc: ConvertFuncToRepresentationFuncAdapter[Model](toRepresentation),
			InternalValueFunc:  ConvertFuncToInternalValueFuncAdapter(toInternalValue),
		}
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

type RepresentationFunc[Model any] func(*models.InternalValue[Model], string) (interface{}, error)
type InternalValueFunc func(map[string]interface{}, string) (interface{}, error)

func ConvertFuncToRepresentationFuncAdapter[Model any](cf types.ConvertFunc) RepresentationFunc[Model] {
	return func(intVal *models.InternalValue[Model], name string) (interface{}, error) {
		return cf(intVal.Map[name])
	}
}

func ConvertFuncToInternalValueFuncAdapter(cf types.ConvertFunc) InternalValueFunc {
	return func(reprModel map[string]interface{}, name string) (interface{}, error) {
		return cf(reprModel[name])
	}
}

func RepresentationPassthrough[Model any]() RepresentationFunc[Model] {
	return func(intVal *models.InternalValue[Model], name string) (interface{}, error) {
		return intVal.Map[name], nil
	}
}

func InternalValuePassthrough() InternalValueFunc {
	return func(reprModel map[string]interface{}, name string) (interface{}, error) {
		return reprModel[name], nil
	}
}

type Field[Model any] struct {
	ItsName            string
	RepresentationFunc RepresentationFunc[Model]
	InternalValueFunc  InternalValueFunc
	Readable           bool
	Writable           bool
}

func (s *Field[Model]) Name() string {
	return s.ItsName
}

func (s *Field[Model]) ToRepresentation(intVal *models.InternalValue[Model]) (interface{}, error) {
	return s.RepresentationFunc(intVal, s.ItsName)
}

func (s *Field[Model]) ToInternalValue(reprModel map[string]interface{}) (interface{}, error) {
	return s.InternalValueFunc(reprModel, s.ItsName)
}

func ExistingField[Model any](name string) Field[Model] {
	return Field[Model]{
		ItsName: name,
		RepresentationFunc: func(intVal *models.InternalValue[Model], name string) (interface{}, error) {
			return intVal.Map[name], nil
		},
		InternalValueFunc: func(reprModel map[string]interface{}, name string) (interface{}, error) {
			return reprModel[name], nil
		},
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
