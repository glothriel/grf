package serializers

import (
	"reflect"
	"strings"

	"github.com/glothriel/gin-rest-framework/pkg/models"
	playgroundValidate "github.com/go-playground/validator/v10"
	gookitValidate "github.com/gookit/validate"
	"github.com/sirupsen/logrus"
)

type ValidatingSerializer[Model any] struct {
	child      Serializer[Model]
	Validators []Validator[Model]
}

func (s *ValidatingSerializer[Model]) ToInternalValue(raw map[string]interface{}) (*models.InternalValue[Model], error) {
	intVal, err := s.child.ToInternalValue(raw)
	if err != nil {
		return intVal, err
	}
	return intVal, s.Validate(intVal)
}

func (s *ValidatingSerializer[Model]) ToRepresentation(intVal *models.InternalValue[Model]) (map[string]interface{}, error) {
	return s.child.ToRepresentation(intVal)
}

func (s *ValidatingSerializer[Model]) FromDB(raw map[string]interface{}) (*models.InternalValue[Model], error) {
	intVal, err := s.child.FromDB(raw)
	if err != nil {
		return intVal, err
	}
	return intVal, s.Validate(intVal)
}

func (s *ValidatingSerializer[Model]) Validate(intVal *models.InternalValue[Model]) error {
	errors := make([]error, 0)
	for _, validator := range s.Validators {
		err := validator.Validate(intVal)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (s *ValidatingSerializer[Model]) WithValidator(validator Validator[Model]) *ValidatingSerializer[Model] {
	s.Validators = append(s.Validators, validator)
	return s
}

func NewValidatingSerializer[Model any](child Serializer[Model]) *ValidatingSerializer[Model] {
	return &ValidatingSerializer[Model]{child: child}
}

type Validator[Model any] interface {
	Validate(*models.InternalValue[Model]) error
}

type GookitValidator[Model any] struct {
	Validation *gookitValidate.Validation
	Rules      []GookitRule
}

type GookitRule struct {
	Fields    string
	Validator string
	Args      []interface{}
}

func (v *GookitValidator[Model]) Validate(intVal *models.InternalValue[Model]) error {
	entity, asModelErr := intVal.AsModel()
	if asModelErr != nil {
		return asModelErr
	}
	val := gookitValidate.New(entity)
	for _, rule := range v.Rules {
		val.AddRule(rule.Fields, rule.Validator, rule.Args...)
	}
	return val.ValidateE().ErrOrNil()
}

func (v *GookitValidator[Model]) WithRule(Fields, Validator string, Args ...interface{}) *GookitValidator[Model] {
	v.Rules = append(v.Rules, GookitRule{Fields, Validator, Args})
	return v
}

func NewGoogkitValidator[Model any]() *GookitValidator[Model] {
	return &GookitValidator[Model]{Validation: gookitValidate.New(nil)}
}

type GoPlaygroundValidator[Model any] struct{}

func (v *GoPlaygroundValidator[Model]) Validate(intVal *models.InternalValue[Model]) error {

	entity, asModelErr := intVal.AsModel()
	if asModelErr != nil {
		return asModelErr
	}
	validator := playgroundValidate.New()
	validator.RegisterTagNameFunc(
		func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		},
	)
	upstreamErr := validator.Struct(entity)
	if upstreamErr == nil {
		return nil
	}
	violations, ok := upstreamErr.(playgroundValidate.ValidationErrors)
	if !ok {
		logrus.Error("Unexpected error type returned from validator")
		return upstreamErr
	}
	validationErr := &ValidationError{FieldErrors: make(map[string][]string)}
	for _, violation := range violations {
		_, ok := validationErr.FieldErrors[violation.Field()]
		if !ok {
			validationErr.FieldErrors[violation.Field()] = make([]string, 0)
		}
		validationErr.FieldErrors[violation.Field()] = append(validationErr.FieldErrors[violation.Field()], violation.Error())
	}
	return validationErr
}

type ValidationError struct {
	FieldErrors map[string][]string
}

// Uses string builder to build error message
func (e *ValidationError) Error() string {
	var sb strings.Builder
	for field, err := range e.FieldErrors {
		sb.WriteString(field)
		sb.WriteString(": ")
		for _, msg := range err {
			sb.WriteString(msg)
			sb.WriteString(", ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
