package serializers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	playgroundValidate "github.com/go-playground/validator/v10"
)

type ValidatingSerializer[Model any] struct {
	child      Serializer
	validators []Validator
}

func (s *ValidatingSerializer[Model]) ToInternalValue(raw map[string]any, ctx *gin.Context) (models.InternalValue, error) {
	intVal, err := s.child.ToInternalValue(raw, ctx)
	if err != nil {
		return intVal, err
	}
	return intVal, s.validate(intVal, ctx)
}

func (s *ValidatingSerializer[Model]) ToRepresentation(intVal models.InternalValue, ctx *gin.Context) (Representation, error) {
	return s.child.ToRepresentation(intVal, ctx)
}

func (s *ValidatingSerializer[Model]) FromDB(raw map[string]any, ctx *gin.Context) (models.InternalValue, error) {
	return s.child.FromDB(raw, ctx)
}

func (s *ValidatingSerializer[Model]) validate(intVal models.InternalValue, ctx *gin.Context) error {
	errors := make([]error, 0)
	for _, validator := range s.validators {
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

func (s *ValidatingSerializer[Model]) AddValidator(validator Validator) *ValidatingSerializer[Model] {
	s.validators = append(s.validators, validator)
	return s
}

func NewValidatingSerializer[Model any](child Serializer, validator ...Validator) *ValidatingSerializer[Model] {
	s := &ValidatingSerializer[Model]{child: child}

	for _, v := range validator {
		s.AddValidator(v)
	}

	return s
}

type Validator interface {
	Validate(models.InternalValue) error
}

type goPlaygroundValidator[Model any] struct {
	rules map[string]any
}

func (v *goPlaygroundValidator[Model]) Validate(intVal models.InternalValue) (err error) {

	validator := playgroundValidate.New()
	validationErrorsByFieldName := validator.ValidateMap(intVal, v.rules)
	validationErr := &ValidationError{FieldErrors: make(map[string][]string)}
	for fieldName, violation := range validationErrorsByFieldName {
		// For some reason ValidateMap includes an empty field name, replace it with the actual field name
		validationErr.FieldErrors[fieldName] = []string{
			strings.Replace(
				violation.(playgroundValidate.ValidationErrors).Error(),
				"''",
				fmt.Sprintf("'%s'", fieldName),
				-1,
			),
		}
	}
	if len(validationErr.FieldErrors) == 0 {
		return nil
	}
	return validationErr
}

func NewGoPlaygroundValidator[Model any](
	rules map[string]any,
) *goPlaygroundValidator[Model] {
	return &goPlaygroundValidator[Model]{
		rules: rules,
	}
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
