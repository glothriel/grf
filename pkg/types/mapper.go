package types

import (
	"fmt"
	"time"
)

type FieldType struct {
	InternalToResponse ConvertFunc
	RequestToInternal  ConvertFunc
	DBToInternal       ConvertFunc
}

type FieldTypeMapper struct {
	Registered map[string]FieldType
}

func (s *FieldTypeMapper) ToRepresentation(typeString string) (ConvertFunc, error) {
	if fieldType, ok := s.Registered[typeString]; ok {
		return func(i any) (any, error) {
			result, resultErr := fieldType.InternalToResponse(i)
			if resultErr != nil {
				return nil, fmt.Errorf(
					"Error converting internal value to representation for type `%s`: %s",
					typeString, resultErr.Error(),
				)
			}
			return result, nil
		}, nil
	}
	return nil, fmt.Errorf("No representation function registered for type `%s`", typeString)
}

func (s *FieldTypeMapper) ToInternalValue(typeString string) (ConvertFunc, error) {
	if fieldType, ok := s.Registered[typeString]; ok {
		return func(i any) (any, error) {
			result, resultErr := fieldType.RequestToInternal(i)
			if resultErr != nil {
				return nil, fmt.Errorf(
					"Error converting request value to internal value for type `%s`: %s",
					typeString, resultErr.Error(),
				)
			}
			return result, nil
		}, nil
	}
	return nil, fmt.Errorf("No internal value function registered for type `%s`", typeString)
}

func (s *FieldTypeMapper) Register(typeString string, fieldType FieldType) {
	s.Registered[typeString] = fieldType
}

func DefaultFieldTypeMapper() *FieldTypeMapper {
	registered := make(map[string]FieldType)
	// all the types that can be directly decoded to and from JSON registered as passthrough
	registered["string"] = FieldType{
		InternalToResponse: ConvertPassThroughWithTypeValidation[string],
		RequestToInternal:  ConvertPassThroughWithTypeValidation[string],
	}
	registered["int"] = FieldType{
		InternalToResponse: ConvertPassThroughWithTypeValidation[int],
		RequestToInternal:  ConvertPassThroughWithTypeValidation[int],
	}
	registered["float64"] = FieldType{
		InternalToResponse: ConvertPassThroughWithTypeValidation[float64],
		RequestToInternal:  ConvertPassThroughWithTypeValidation[float64],
	}
	registered["bool"] = FieldType{
		InternalToResponse: ConvertPassThroughWithTypeValidation[bool],
		RequestToInternal:  ConvertPassThroughWithTypeValidation[bool],
	}
	registered["time.Time"] = FieldType{
		InternalToResponse: ConvertPassThroughWithTypeValidation[time.Time],
		RequestToInternal:  ConvertPassThroughWithTypeValidation[time.Time],
	}
	return &FieldTypeMapper{
		Registered: registered,
	}
}

type ConvertFunc func(any) (any, error)

func ConvertPassThrough(in any) (any, error) {
	return in, nil
}

func ConvertPassThroughWithTypeValidation[T any](in any) (any, error) {
	if _, ok := in.(T); !ok {
		var t T
		return nil, fmt.Errorf("Expected type `%T`, got `%T`", t, in)
	}
	return in, nil
}
