package serializers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/detectors"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/sirupsen/logrus"
)

type ModelSerializer[Model any] struct {
	Fields map[string]*fields.Field[Model]

	toRepresentationDetector detectors.ToRepresentationDetector[Model]
	toInternalValueDetector  detectors.ToInternalValueDetector
}

func (s *ModelSerializer[Model]) ToInternalValue(raw map[string]any, ctx *gin.Context) (models.InternalValue, error) {
	intVMap := make(map[string]any)
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
		intV, err := field.ToInternalValue(raw, ctx)
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
	return intVMap, nil
}

func (s *ModelSerializer[Model]) ToRepresentation(intVal models.InternalValue, ctx *gin.Context) (Representation, error) {
	raw := make(map[string]any)
	for _, field := range s.Fields {
		if !field.Readable {
			continue
		}
		value, err := field.ToRepresentation(intVal, ctx)

		if err != nil {
			return nil, fmt.Errorf(
				"Failed to serialize field `%s` to representation: %w", field.Name(), err,
			)
		}
		raw[field.Name()] = value
	}
	return raw, nil
}

func (s *ModelSerializer[Model]) FromDB(raw map[string]any, ctx *gin.Context) (models.InternalValue, error) {
	intVMap := make(models.InternalValue)
	for k := range raw {
		field, ok := s.Fields[k]
		if !ok {
			continue
		}
		intV, err := field.FromDB(raw, ctx)
		if err != nil {
			return nil, fmt.Errorf("Failed to serialize field `%s` from database: %s", k, err)
		}
		intVMap[k] = intV
	}

	return intVMap, nil
}

func (s *ModelSerializer[Model]) Validate(intVal models.InternalValue, ctx *gin.Context) error {
	return nil
}

func (s *ModelSerializer[Model]) WithNewField(field *fields.Field[Model]) *ModelSerializer[Model] {
	s.Fields[field.Name()] = field
	return s
}

func (s *ModelSerializer[Model]) WithField(name string, updateFunc func(oldField *fields.Field[Model])) *ModelSerializer[Model] {
	v, ok := s.Fields[name]
	if !ok {
		var m Model
		logrus.Panicf("Could not find field `%s` on model `%s` when registering serializer field", name, reflect.TypeOf(m))
	}
	updateFunc(v)
	return s
}

func (s *ModelSerializer[Model]) WithModelFields(passedFields []string) *ModelSerializer[Model] {

	s.Fields = make(map[string]*fields.Field[Model])
	var m Model
	for _, field := range passedFields {
		toRepresentation, toRepresentationErr := s.toRepresentationDetector.ToRepresentation(field)
		if toRepresentationErr != nil {
			logrus.Panicf("WithModelFields: Failed to register model `%s` fields: %s", reflect.TypeOf(m), toRepresentationErr)

		}
		toInternalValue, toInternalValueErr := s.toInternalValueDetector.ToInternalValue(field)
		if toInternalValueErr != nil {
			logrus.Panicf("WithModelFields: Failed to register model `%s` fields: %s", reflect.TypeOf(m), toInternalValueErr)
		}
		s.Fields[field] = fields.NewField[Model](
			field,
		).WithRepresentationFunc(
			toRepresentation,
		).WithInternalValueFunc(
			toInternalValue,
		)
	}
	return s
}

func NewModelSerializer[Model any]() *ModelSerializer[Model] {
	return (&ModelSerializer[Model]{
		toRepresentationDetector: detectors.DefaultToRepresentationDetector[Model](),
		toInternalValueDetector:  detectors.DefaultToInternalValueDetector[Model](),
	}).WithModelFields(detectors.Fields[Model]())
}
