package serializers

import (
	"fmt"
	"reflect"

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

	for k := range s.Fields {
		field, ok := s.Fields[k]
		if !ok {
			superfluousFields = append(superfluousFields, k)
			continue
		}
		if !field.Writable {
			continue
		}
		// Please remember, that `ToInteralValue` doesn't necessarily extract the value from the `raw` map.
		// In theory it could use request headers, cookies, external APIs, queries or anything else.
		intV, err := field.ToInternalValue(raw, ctx)
		if err != nil {
			_, isMissingFieldErr := err.(fields.ErrorFieldIsNotPresentInPayload)
			if isMissingFieldErr {
				continue
			}
			return nil, &ValidationError{FieldErrors: map[string][]string{k: {err.Error()}}}
		}
		intVMap[k] = intV
	}
	if len(superfluousFields) > 0 {
		errMap := map[string][]string{}
		for _, field := range superfluousFields {
			errMap[field] = []string{fmt.Sprintf("Field `%s` is not accepted by this endpoint", field)}
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
			_, isMissingFieldErr := err.(fields.ErrorFieldIsNotPresentInPayload)
			if isMissingFieldErr {
				continue
			}
			return nil, &ValidationError{
				FieldErrors: map[string][]string{
					field.Name(): {
						err.Error(),
					},
				},
			}
		}
		raw[field.Name()] = value
	}
	return raw, nil
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
			if toRepresentationErr == detectors.ErrFieldShouldBeSkipped {
				logrus.Infof("WithModelFields: Skipping model `%s` field `%s`", reflect.TypeOf(m), field)
				continue
			}
			logrus.Panicf("WithModelFields: Failed to register model `%s` fields: %s", reflect.TypeOf(m), toRepresentationErr)

		}
		toInternalValue, toInternalValueErr := s.toInternalValueDetector.ToInternalValue(field)
		if toInternalValueErr != nil {
			if toRepresentationErr == detectors.ErrFieldShouldBeSkipped {
				logrus.Infof("WithModelFields: Skipping model `%s` field `%s`", reflect.TypeOf(m), field)
				continue
			}
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
	}).WithModelFields(
		detectors.Fields[Model](),
	).WithField("id", func(oldField *fields.Field[Model]) { oldField.ReadOnly() })
}
