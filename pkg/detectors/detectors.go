package detectors

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/types"
)

// ToRepresentationDetector is an interface that allows to detect the representation function for a given field
// using well known interfaces, knowledge of stdlib types or any other way
type ToRepresentationDetector[Model any] interface {
	ToRepresentation(fieldName string) (fields.RepresentationFunc, error)
}

func DefaultToRepresentationDetector[Model any]() ToRepresentationDetector[Model] {
	return NewChainingToRepresentationDetector[Model](
		NewUsingGRFRepresentableToRepresentationProvider[Model](),
		NewFromTypeMapperToRepresentationProvider[Model](types.Mapper()),
		NewEncodingTextMarshalerToRepresentationProvider[Model](),
	)
}

type usingGRFRepresentableToRepresentationDetector[Model any] struct{}

func (p *usingGRFRepresentableToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isGRFRepresentable {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				return v.(fields.GRFRepresentable).ToRepresentation()
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a GRFRepresentable", fieldName)
}

func NewUsingGRFRepresentableToRepresentationProvider[Model any]() ToRepresentationDetector[Model] {
	return &usingGRFRepresentableToRepresentationDetector[Model]{}
}

// fromTypeMapperToRepresentationDetector uses field type mapper to detect the representation function for a given field
type fromTypeMapperToRepresentationDetector[Model any] struct {
	mapper         *types.FieldTypeMapper
	modelTypeNames map[string]string
}

func (p *fromTypeMapperToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	ftmToRepresentation, toRepresentationErr := p.mapper.ToRepresentation(p.modelTypeNames[fieldName])
	if toRepresentationErr != nil {
		return nil, toRepresentationErr
	}
	return ConvertFuncToRepresentationFuncAdapter(ftmToRepresentation), nil
}

func NewFromTypeMapperToRepresentationProvider[Model any](mapper *types.FieldTypeMapper) ToRepresentationDetector[Model] {
	return &fromTypeMapperToRepresentationDetector[Model]{
		mapper:         mapper,
		modelTypeNames: FieldTypes[Model](),
	}
}

type encodingTextMarshalerToRepresentationDetector[Model any] struct{}

func (p *encodingTextMarshalerToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isEncodingTextMarshaler {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				marshalledBytes, marshallErr := v.(encoding.TextMarshaler).MarshalText()
				if marshallErr != nil {
					return nil, marshallErr
				}
				return string(marshalledBytes), nil
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextMarshaler", fieldName)
}

func NewEncodingTextMarshalerToRepresentationProvider[Model any]() ToRepresentationDetector[Model] {
	return &encodingTextMarshalerToRepresentationDetector[Model]{}
}

type chaningToRepresentationDetector[Model any] struct {
	children []ToRepresentationDetector[Model]
}

func (p *chaningToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	for _, child := range p.children {
		representation, representationErr := child.ToRepresentation(fieldName)
		if representationErr == nil {
			return representation, nil
		}
	}
	var m Model
	return nil, fmt.Errorf("No representation function could be found for field `%T`.`%s`", m, fieldName)
}

func NewChainingToRepresentationDetector[Model any](children ...ToRepresentationDetector[Model]) ToRepresentationDetector[Model] {
	return &chaningToRepresentationDetector[Model]{children: children}
}

type ToInternalValueDetector interface {
	ToInternalValue(fieldName string) (fields.InternalValueFunc, error)
}

func DefaultToInternalValueDetector[Model any]() ToInternalValueDetector {
	return NewChainingToInternalValueDetector[Model](
		NewUsingGRFParsableToInternalValueProvider[Model](),
		NewFromTypeMapperToInternalValueProvider[Model](types.Mapper()),
		NewEncodingTextUnmarshalerToInternalValueProvider[Model](),
	)
}

type usingGRFParsableToInternalValueDetector[Model any] struct{}

func (p *usingGRFParsableToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isGRFParsable {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				typedV := reflect.New(fieldSettings.itsType).Interface()
				theErr := typedV.(fields.GRFParsable).FromRepresentation(v)
				return typedV, theErr
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a GRFParsable", fieldName)
}

func NewUsingGRFParsableToInternalValueProvider[Model any]() ToInternalValueDetector {
	return &usingGRFParsableToInternalValueDetector[Model]{}
}

type fromTypeMapperToInternalValueDetector[Model any] struct {
	mapper         *types.FieldTypeMapper
	modelTypeNames map[string]string
}

func (p *fromTypeMapperToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	ftmToInternalValue, toInternalValueErr := p.mapper.ToInternalValue(p.modelTypeNames[fieldName])
	if toInternalValueErr != nil {
		return nil, toInternalValueErr
	}
	return ConvertFuncToInternalValueFuncAdapter(ftmToInternalValue), nil
}

func NewFromTypeMapperToInternalValueProvider[Model any](mapper *types.FieldTypeMapper) ToInternalValueDetector {
	return &fromTypeMapperToInternalValueDetector[Model]{mapper: mapper, modelTypeNames: FieldTypes[Model]()}
}

type encodingTextUnmarshalerToInternalValueDetector[Model any] struct{}

func (p *encodingTextUnmarshalerToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isEncodingTextUnmarshaler {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				var realFieldValue any = reflect.New(fieldSettings.itsType).Interface()
				vStr, ok := v.(string)
				if ok {
					if fieldSettings.isEncodingTextUnmarshaler {
						fv := realFieldValue.(encoding.TextUnmarshaler)
						unmarshalErr := fv.UnmarshalText([]byte(vStr))
						return fv, unmarshalErr
					} else {
						return types.ConvertPassThrough(v)
					}
				}
				return nil, fmt.Errorf("Field `%s` is not a string", fieldName)
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextUnmarshaler", fieldName)
}

func NewEncodingTextUnmarshalerToInternalValueProvider[Model any]() ToInternalValueDetector {
	return &encodingTextUnmarshalerToInternalValueDetector[Model]{}
}

type chaningToInternalValueDetector[Model any] struct {
	children []ToInternalValueDetector
}

func (p *chaningToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	for _, child := range p.children {
		internalValue, internalValueErr := child.ToInternalValue(fieldName)
		if internalValueErr == nil {
			return internalValue, nil
		}
	}
	var m Model
	return nil, fmt.Errorf("No internal value function could be found for field `%T`.`%s`", m, fieldName)
}

func NewChainingToInternalValueDetector[Model any](children ...ToInternalValueDetector) ToInternalValueDetector {
	return &chaningToInternalValueDetector[Model]{children: children}
}
