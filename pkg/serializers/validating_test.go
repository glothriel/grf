package serializers

import (
	"errors"
	"testing"

	"github.com/glothriel/grf/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockValidatedModel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Age       int    `json:"age"`
	IsMarried bool   `json:"is_married"`
}

func TestGoPlaygroundValidator(t *testing.T) {
	// given
	validator := NewGoPlaygroundValidator[mockValidatedModel](
		map[string]any{
			"name":       "required",
			"surname":    "required",
			"age":        "required,gt=0,lt=130",
			"is_married": "required",
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"name":       "John",
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	})

	// then
	assert.Nil(t, err)
}

func TestGoPlaygroundValidatorErr(t *testing.T) {
	// given
	validator := NewGoPlaygroundValidator[mockValidatedModel](
		map[string]any{
			"age": "required,gt=0,lt=130",
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"age": 2000,
	})

	// then
	assert.Error(t, err)
}

func TestValidatingSerializerToInternalValue(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
	)

	// when
	intVal, err := serializer.ToInternalValue(map[string]any{
		"name":       "John",
		"surname":    "Doe",
		"age":        20.0,
		"is_married": true,
	}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, models.InternalValue{
		"name":       "John",
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	}, intVal)
}

func TestValidatingSerializerToInternalValueErr(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
	)

	// when
	_, err := serializer.ToInternalValue(map[string]any{
		"name":       "John",
		"surname":    "Doe",
		"age":        "foo",
		"is_married": true,
	}, nil)

	// then
	assert.Error(t, err)
}

func TestValidatingSerializerToRepresentation(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
	)

	// when
	rep, err := serializer.ToRepresentation(models.InternalValue{
		"name":       "John",
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, Representation{
		"name":       "John",
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	}, rep)
}

type mockValidator struct {
	shouldFail bool
}

func (v *mockValidator) Validate(models.InternalValue) error {
	if v.shouldFail {
		return errors.New("validation failed")
	}
	return nil
}

func TestValidatingSerializerValidate(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
		&mockValidator{},
	)

	// when
	err := serializer.validate(models.InternalValue{}, nil)

	// then
	assert.NoError(t, err)
}

func TestValidatingSerializerValidateErr(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
		&mockValidator{shouldFail: true},
	)

	// when
	err := serializer.validate(models.InternalValue{}, nil)

	// then
	assert.Error(t, err)
}

func TestValidatingSerializerAddValidator(t *testing.T) {
	// given
	serializer := NewValidatingSerializer[mockValidatedModel](
		NewModelSerializer[mockValidatedModel](),
	)

	// when
	serializer.AddValidator(&mockValidator{})

	// then
	assert.Len(t, serializer.validators, 1)
}

func TestJSONSchemaValidator(t *testing.T) {
	// given
	validator := NewJSONSchemaValidator(
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
				"age": map[string]any{
					"type": "number",
				},
			},
			"required": []string{"name"},
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"name":       "John",
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	})

	// then
	assert.Nil(t, err)
}

func TestJSONSchemaValidationSchemaLvlErr(t *testing.T) {
	// given
	validator := NewJSONSchemaValidator(
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
				"age": map[string]any{
					"type": "number",
				},
			},
			"required": []string{"name"},
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"surname":    "Doe",
		"age":        20,
		"is_married": true,
	})

	// then
	validationErr := err.(*ValidationError)
	assert.Equal(t, validationErr.FieldErrors["all"], []string{
		"missing properties: 'name'",
	})
}

func TestJSONSchemaValidationAttributeLvlErr(t *testing.T) {
	// given
	validator := NewJSONSchemaValidator(
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
				"age": map[string]any{
					"type": "number",
				},
				"status": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"active": map[string]any{
							"type": "boolean",
						},
					},
					"required": []string{"active"},
				},
			},
			"required": []string{"name"},
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"name":       "John",
		"surname":    "Doe",
		"age":        "huehue not a nubmer",
		"is_married": true,
		"status": map[string]any{
			"active": ":-)",
		},
	})

	// then
	validationErr := err.(*ValidationError)
	logrus.Error(validationErr.FieldErrors)
	assert.Equal(t, validationErr.FieldErrors["age"], []string{
		"expected number, but got string",
	})
	assert.Equal(t, validationErr.FieldErrors["status.active"], []string{
		"expected boolean, but got string",
	})
}

func TestSimpleValidator(t *testing.T) {
	// given
	validator := NewSimpleValidator(
		func(iv models.InternalValue) error {
			if iv["age"].(int) < 18 {
				return errors.New("age must be at least 18")
			}
			return nil
		},
	)

	// when
	err := validator.Validate(map[string]any{
		"name": "John",
		"age":  20,
	})

	// then
	assert.Nil(t, err)

	// when
	err = validator.Validate(map[string]any{
		"name": "Henry",
		"age":  17,
	})

	// then
	assert.Error(t, err)
}
