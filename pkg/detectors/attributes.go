package detectors

import (
	"encoding"
	"reflect"

	"github.com/glothriel/grf/pkg/fields"
)

// Prints a summary with the fields of the model obtained using reflection
func FieldTypes[Model any]() map[string]string {
	ret := make(map[string]string)

	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous {
			ret[field.Tag.Get("json")] = field.Type.String()
		}
	}
	return ret
}

func Fields[Model any]() []string {
	fieldNames := []string{}
	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous {
			fieldNames = append(fieldNames, field.Tag.Get("json"))
		}
	}
	return fieldNames
}

type fieldSettings struct {
	itsType                   reflect.Type
	isEncodingTextMarshaler   bool
	isEncodingTextUnmarshaler bool

	isGRFRepresentable bool
	isGRFParsable      bool
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
			_, isGRFRepresentable := theTypeAsAny.(fields.GRFRepresentable)
			_, isGRFParsable := theTypeAsAny.(fields.GRFParsable)

			settings = &fieldSettings{
				itsType: reflect.TypeOf(
					reflectedInstance.Interface(),
				),
				isEncodingTextMarshaler:   isEncodingTextMarshaler,
				isEncodingTextUnmarshaler: isEncodingTextUnmarshaler,
				isGRFRepresentable:        isGRFRepresentable,
				isGRFParsable:             isGRFParsable,
			}
		}
	}
	return settings
}
