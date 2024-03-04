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
	registered["int"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return int(f)
			},
			true,
		),
	}
	registered["int8"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return int8(f)
			},
			true,
		),
	}

	registered["int16"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return int16(f)
			},
			true,
		),
	}
	registered["int32"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return int32(f)
			},
			true,
		),
	}
	registered["int64"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return int64(f)
			},
			true,
		),
	}
	registered["uint"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return uint(f)
			},
			false,
		),
	}
	registered["uint8"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return uint8(f)
			},
			false,
		),
	}
	registered["uint16"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return uint16(f)
			},
			false,
		),
	}
	registered["uint32"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return uint32(f)
			},
			false,
		),
	}
	registered["uint64"] = FieldType{
		InternalToResponse: ConvertPassThrough,
		RequestToInternal: ConvertFloatDynamic(
			func(f float64) any {
				return uint64(f)
			},
			false,
		),
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

func ConvertFloatDynamic(convert func(float64) any, canBeBelowZero bool) func(any) (any, error) {
	return func(in any) (any, error) {
		if f, ok := in.(float64); ok {
			if math.Mod(f, 1) == 0 && ((!canBeBelowZero && f >= 0) || canBeBelowZero) {
				return convert(f), nil
			} else {
				return 0, fmt.Errorf("Value %f is not an integer", f)
			}
		}
		return nil, fmt.Errorf("Expected type `float64`, got `%T`", in)
	}
}
