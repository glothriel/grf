package types

import (
	"fmt"
	"math"
	"time"
)

var NumericTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
}
var mapper *FieldTypeMapper

func Mapper() *FieldTypeMapper {
	if mapper == nil {
		mapper = DefaultFieldTypeMapper()
	}
	return mapper
}

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
	// int is a special case, because JSON only has float64, so we need to convert
	for _, t := range []string{"int", "int8", "int16", "int32", "int64"} {
		registered[t] = FieldType{
			InternalToResponse: ConvertPassThrough,
			RequestToInternal:  ConvertFloatToInt,
		}
	}
	for _, t := range []string{"uint", "uint8", "uint16", "uint32", "uint64"} {
		registered[t] = FieldType{
			InternalToResponse: ConvertPassThrough,
			RequestToInternal:  ConvertFloatToUint,
		}
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

func ConvertFloatToInt(in any) (any, error) {
	if f, ok := in.(float64); ok {
		if math.Mod(f, 1) == 0 {
			i := int(f)
			return i, nil
		} else {
			return 0, fmt.Errorf("Value %f is not an integer", f)
		}
	}
	return nil, fmt.Errorf("Expected type `float64`, got `%T`", in)
}

func ConvertFloatToUint(in any) (any, error) {
	if f, ok := in.(float64); ok {
		if math.Mod(f, 1) == 0 && f >= 0 {
			i := int(f)
			return i, nil
		} else {
			return 0, fmt.Errorf("Value %f is not an unsigned integer", f)
		}
	}
	return nil, fmt.Errorf("Expected type `float64`, got `%T`", in)
}
