package serializers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	playgroundValidate "github.com/go-playground/validator/v10"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
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

type simpleValidator struct {
	validateFunc func(models.InternalValue) error
}

func (v *simpleValidator) Validate(intVal models.InternalValue) error {
	return v.validateFunc(intVal)
}

func NewSimpleValidator(validateFunc func(models.InternalValue) error) Validator {
	return &simpleValidator{validateFunc: validateFunc}
}

type jsonSchemaValidator struct {
	schema *jsonschema.Schema
}

// TransformError recursively transforms a ValidationError into a map[string][]string.
func TransformError(err *jsonschema.ValidationError) map[string][]string {

	result := make(map[string][]string)

	// Helper function to add messages to the map.
	addMessage := func(key, message string) {
		if _, exists := result[key]; !exists {
			result[key] = []string{}
		}
		result[key] = append(result[key], message)
	}

	// Recursive function to process the ValidationError.
	var processError func(err *jsonschema.ValidationError)
	processError = func(err *jsonschema.ValidationError) {
		if len(err.Causes) == 0 {
			// It's a leaf error, add its message to the map.
			location := err.InstanceLocation
			if location == "" {
				location = "all"
			} else {
				location = strings.Replace(location, "/", ".", -1)
				location = location[1:]
			}
			addMessage(location, err.Message)
		} else {
			// Process nested errors.
			for _, cause := range err.Causes {
				processError(cause)
			}
		}
	}

	processError(err)
	return result
}

func (v *jsonSchemaValidator) Validate(intVal models.InternalValue) error {

	if validateErr := v.schema.Validate(
		map[string]any(intVal),
	); validateErr != nil {
		jsonSchemaValidationErr, ok := validateErr.(*jsonschema.ValidationError)
		if !ok {
			return &ValidationError{
				FieldErrors: map[string][]string{
					"all": {validateErr.Error()},
				},
			}
		}
		// Convert the jsonschema.ValidationError to a ValidationError
		fieldErrors := TransformError(jsonSchemaValidationErr)

		return &ValidationError{
			FieldErrors: fieldErrors,
		}
	}

	return nil
}

func NewJSONSchemaValidator(rawSchema map[string]any) Validator {
	rawSchema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	rawSchema["$id"] = "https://glothriel.github.io/grf/schema.json"
	encodedSchema, marshalErr := json.Marshal(rawSchema)
	if marshalErr != nil {
		logrus.Panicf("Error marshaling JSONSchema: %s", marshalErr)
	}
	compiledSchema, compileErr := jsonschema.CompileString("schema.json", string(encodedSchema))
	if compileErr != nil {
		logrus.Panicf("Error compiling JSONSchema: %s", compileErr)
	}
	return &jsonSchemaValidator{schema: compiledSchema}
}
