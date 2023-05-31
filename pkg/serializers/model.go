package serializers

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/types"
	"github.com/sirupsen/logrus"
)

type ModelSerializer[Model any] struct {
	Fields          map[string]*fields.Field[Model]
	FieldTypeMapper *types.FieldTypeMapper
}

func (s *ModelSerializer[Model]) ToInternalValue(raw map[string]any, ctx *grfctx.Context) (models.InternalValue, error) {
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
		intV, err := field.InternalValueFunc(raw, k, ctx)
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

func (s *ModelSerializer[Model]) ToRepresentation(intVal models.InternalValue, ctx *grfctx.Context) (Representation, error) {
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

func (s *ModelSerializer[Model]) FromDB(raw map[string]any, ctx *grfctx.Context) (models.InternalValue, error) {
	intVMap := make(models.InternalValue)
	for k := range raw {
		field, ok := s.Fields[k]
		if !ok {
			continue
		}
		intV, err := field.FromDBFunc(raw, k, ctx)
		if err != nil {
			return nil, fmt.Errorf("Failed to serialize field `%s` from database: %s", k, err)
		}
		intVMap[k] = intV
	}

	return intVMap, nil
}

func (s *ModelSerializer[Model]) Validate(intVal models.InternalValue, ctx *grfctx.Context) error {
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
		logrus.Fatalf("Could not find field `%s` on model `%s` when registering serializer field", name, reflect.TypeOf(m))
	}
	updateFunc(v)
	return s
}

func (s *ModelSerializer[Model]) WithModelFields(passedFields []string) *ModelSerializer[Model] {
	s.Fields = make(map[string]*fields.Field[Model])
	var m Model
	attributeTypes := DetectAttributes(m)
	for _, field := range passedFields {
		attributeType, ok := attributeTypes[field]
		if !ok {
			logrus.Fatalf("Could not find field `%s` on model `%s` when registering serializer fields", field, reflect.TypeOf(m))
		}
		toRepresentation, toRepresentationErr := s.FieldTypeMapper.ToRepresentation(attributeType)
		if toRepresentationErr != nil {
			logrus.Debugf("Could not determine representation of field `%s` on model `%s`: %s", field, reflect.TypeOf(m), toRepresentationErr)
			toRepresentation = EncodeToStringOrPassthrough[Model](field)
		}
		toInternalValue, toInternalValueErr := s.FieldTypeMapper.ToInternalValue(attributeType)
		if toInternalValueErr != nil {
			logrus.Debugf("Could not determine internal value of field `%s` on model `%s`: %s", field, reflect.TypeOf(m), toInternalValueErr)
			toInternalValue = DecodeFromStringOrPassthrough[Model](field)
		}
		s.Fields[field] = fields.NewField[Model](
			field,
		).WithRepresentationFunc(
			ConvertFuncToRepresentationFuncAdapter[Model](toRepresentation),
		).WithInternalValueFunc(
			ConvertFuncToInternalValueFuncAdapter(toInternalValue),
		)
	}
	modelAttrs := reflect.VisibleFields(reflect.TypeOf(m))
	v := reflect.ValueOf(m)
	for _, attr := range modelAttrs {
		if !attr.Anonymous {
			updater, ok := v.FieldByName(attr.Name).Interface().(FieldUpdater[Model])
			if ok {
				updater.Update(s.Fields[attr.Tag.Get("json")])
				continue
			}
		}
	}
	return s
}

type FieldUpdater[Model any] interface {
	Update(f *fields.Field[Model])
}

func NewModelSerializer[Model any](ftm *types.FieldTypeMapper) *ModelSerializer[Model] {
	if ftm == nil {
		ftm = types.DefaultFieldTypeMapper()
	}
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
	}).WithModelFields(fieldNames)
}

func ConvertFuncToRepresentationFuncAdapter[Model any](cf types.ConvertFunc) fields.RepresentationFunc[Model] {
	return func(intVal models.InternalValue, name string, ctx *grfctx.Context) (any, error) {
		return cf(intVal[name])
	}
}

func ConvertFuncToInternalValueFuncAdapter(cf types.ConvertFunc) fields.InternalValueFunc {
	return func(reprModel map[string]any, name string, ctx *grfctx.Context) (any, error) {
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

func DecodeFromStringOrPassthrough[Model any](fieldName string) types.ConvertFunc {
	settings := getFieldSettings[Model](fieldName)
	if settings == nil {
		var entity Model
		logrus.Fatalf("Could not find field `%s` on model `%s`", fieldName, reflect.TypeOf(entity))
	}
	return func(v any) (any, error) {
		var realFieldValue any = reflect.New(settings.itsType).Interface()
		vStr, ok := v.(string)
		if ok {
			if settings.isEncodingTextUnmarshaler {
				fv := realFieldValue.(encoding.TextUnmarshaler)
				unmarshalErr := fv.UnmarshalText([]byte(vStr))
				return fv, unmarshalErr
			} else {
				return types.ConvertPassThrough(v)
			}
		}

		return types.ConvertPassThrough(v)
	}
}

func EncodeToStringOrPassthrough[Model any](fieldName string) types.ConvertFunc {
	settings := getFieldSettings[Model](fieldName)
	if settings == nil {
		var entity Model
		logrus.Fatalf("Could not find field `%s` on model `%s`", fieldName, reflect.TypeOf(entity))
	}
	return func(v any) (any, error) {
		if settings.isEncodingTextMarshaler {
			marshalledBytes, marshallErr := v.(encoding.TextMarshaler).MarshalText()
			if marshallErr != nil {
				return nil, marshallErr
			}
			return string(marshalledBytes), nil
		}
		return types.ConvertPassThrough(v)
	}
}

type fieldSettings struct {
	itsType                      reflect.Type
	isEncodingTextMarshaler      bool
	isPtrEncodingTextMarshaler   bool
	isEncodingTextUnmarshaler    bool
	isPtrEncodingTextUnmarshaler bool
}

func getFieldSettings[Model any](fieldName string) *fieldSettings {
	var entity Model
	var settings *fieldSettings
	for _, field := range reflect.VisibleFields(reflect.TypeOf(entity)) {
		jsonTag := field.Tag.Get("json")
		if jsonTag == fieldName {
			var theTypeAsAny any
			reflectedInstance := reflect.New(reflect.TypeOf(reflect.ValueOf(entity).FieldByName(field.Name).Interface())).Elem()
			if reflectedInstance.CanAddr() {
				theTypeAsAny = reflectedInstance.Addr().Interface()
			} else {
				theTypeAsAny = reflectedInstance.Interface()
			}

			_, isEncodingTextMarshaler := theTypeAsAny.(encoding.TextMarshaler)
			_, isEncodingTextUnmarshaler := theTypeAsAny.(encoding.TextUnmarshaler)
			settings = &fieldSettings{
				itsType: reflect.TypeOf(
					reflectedInstance.Interface(),
				),
				isEncodingTextMarshaler:   isEncodingTextMarshaler,
				isEncodingTextUnmarshaler: isEncodingTextUnmarshaler,
			}
		}
	}
	return settings
}
