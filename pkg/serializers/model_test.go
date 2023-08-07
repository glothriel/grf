package serializers

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/detectors"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
)

type mockModel struct {
	Foo string `json:"foo"`
}

func TestModelSerializerToInternalValue(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// when
	intVal, err := serializer.ToInternalValue(map[string]any{"foo": "bar"}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, models.InternalValue{"foo": "bar"}, intVal)
}

func TestModelSerializerToInternalValueSuperflousField(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// when
	_, err := serializer.ToInternalValue(map[string]any{"foo": "bar", "baz": "qux"}, nil)

	// then
	assert.Error(t, err)
}

type anotherMockModel struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

func TestModelSerializerToInternalValueNonWritableField(t *testing.T) {
	// given
	serializer := NewModelSerializer[anotherMockModel]().WithField(
		"bar",
		func(oldField *fields.Field[anotherMockModel]) {
			oldField.ReadOnly()
		},
	)

	// when
	intVal, err := serializer.ToInternalValue(map[string]any{"foo": "bar", "bar": "baz"}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, models.InternalValue{"foo": "bar"}, intVal)
}

func TestModelSerializerToInternalValueFieldErr(t *testing.T) {
	// given
	serializer := NewModelSerializer[anotherMockModel]().WithNewField(
		fields.NewField[anotherMockModel]("foo").WithInternalValueFunc(
			func(m map[string]any, s string, ctx *gin.Context) (any, error) {
				return nil, errors.New("foo err")
			},
		),
	)

	// when
	_, err := serializer.ToInternalValue(map[string]any{"foo": "bar"}, nil)

	// then
	assert.ErrorContains(t, err, "foo err")
}

func TestModelSerializerToRepresentation(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// when
	repr, err := serializer.ToRepresentation(models.InternalValue{"foo": "bar"}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, Representation{"foo": "bar"}, repr)
}

func TestModelSerializerToRepresentationNonReadableField(t *testing.T) {
	// given
	serializer := NewModelSerializer[anotherMockModel]().WithField(
		"bar",
		func(oldField *fields.Field[anotherMockModel]) {
			oldField.WriteOnly()
		},
	)

	// when
	repr, err := serializer.ToRepresentation(models.InternalValue{"foo": "bar", "bar": "baz"}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, Representation{"foo": "bar"}, repr)
}

func TestModelSerializerToRepresentationFieldErr(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]().WithNewField(
		fields.NewField[mockModel]("bar").WithRepresentationFunc(
			func(m models.InternalValue, s string, ctx *gin.Context) (any, error) {
				return nil, errors.New("foo err")
			},
		),
	)

	// when
	_, err := serializer.ToRepresentation(models.InternalValue{"bar": "baz", "foo": "123"}, nil)

	// then
	assert.ErrorContains(t, err, "foo err")
}

func TestModelSerializerValidate(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// when
	err := serializer.Validate(models.InternalValue{"foo": "bar"}, nil)

	// then
	assert.NoError(t, err)
}

func TestModelSerializerWithNewField(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// when
	_, barFieldExists := serializer.Fields["bar"]

	// then
	assert.False(t, barFieldExists)

	// when
	serializer.WithNewField(
		fields.NewField[mockModel]("bar").WithInternalValueFunc(
			func(m map[string]any, s string, ctx *gin.Context) (any, error) {
				return "baz", nil
			},
		),
	)

	// then
	_, barFieldExists = serializer.Fields["bar"]
	assert.True(t, barFieldExists)
}

func TestModelSerializerWithField(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]().WithField(
		"foo",
		func(f *fields.Field[mockModel]) {
			f.WithInternalValueFunc(
				func(m map[string]any, s string, ctx *gin.Context) (any, error) {
					return m[s].(string) + " huehue", nil
				},
			)
		},
	)

	// when
	value, err := serializer.ToInternalValue(map[string]any{"foo": "bar"}, nil)

	// then
	assert.NoError(t, err)
	assert.Equal(t, models.InternalValue{"foo": "bar huehue"}, value)
}

func TestModelSerializerWithFieldDoesNotExist(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()

	// then
	assert.Panics(t, func() {
		serializer.WithField(
			"bar",
			func(f *fields.Field[mockModel]) {
				f.WithInternalValueFunc(
					func(m map[string]any, s string, ctx *gin.Context) (any, error) {
						return nil, nil
					},
				)
			},
		)
	})
}

type mockToRepresentationDetector[Model any] struct {
	shouldErr bool
}

func (d *mockToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	if d.shouldErr {
		return nil, errors.New("foo err")
	}
	return detectors.ConvertFuncToRepresentationFuncAdapter(
		func(v any) (any, error) {
			return "huehue", nil
		},
	), nil
}

type mockToInternalValueDetector[Model any] struct {
	shouldErr bool
}

func (d *mockToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	if d.shouldErr {
		return nil, errors.New("foo err")
	}
	return detectors.ConvertFuncToInternalValueFuncAdapter(
		func(v any) (any, error) {
			return "huehue", nil
		},
	), nil
}

func TestWithModelFields(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()
	serializer.toInternalValueDetector = &mockToInternalValueDetector[mockModel]{}
	serializer.toRepresentationDetector = &mockToRepresentationDetector[mockModel]{}

	// when
	serializer.WithModelFields([]string{"foo"})
	internalValue, internalValueErr := serializer.Fields["foo"].ToInternalValue(map[string]any{"foo": "bar"}, nil)
	representation, representationErr := serializer.Fields["foo"].ToRepresentation(map[string]any{"foo": "bar"}, nil)

	// then
	assert.NoError(t, internalValueErr)
	assert.Equal(t, "huehue", internalValue)
	assert.NoError(t, representationErr)
	assert.Equal(t, "huehue", representation)
}

func TestWithModelFieldsPanicsWhenInternalValueDetectorErr(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()
	serializer.toInternalValueDetector = &mockToInternalValueDetector[mockModel]{shouldErr: true}

	// then
	assert.Panics(t, func() {
		serializer.WithModelFields([]string{"foo"})
	})
}

func TestWithModelFieldsPanicsWhenRepresentationDetectorErr(t *testing.T) {
	// given
	serializer := NewModelSerializer[mockModel]()
	serializer.toRepresentationDetector = &mockToRepresentationDetector[mockModel]{shouldErr: true}

	// then
	assert.Panics(t, func() {
		serializer.WithModelFields([]string{"foo"})
	})
}
