package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockModel struct {
	SomeItems []any `json:"some_items" gorm:"type:text"`
}

func InternalValueCase[Model any, TestedType any](t *testing.T, passedValue any, expectedValue any) {
	baseField := NewField[Model]("some_items")
	field := SliceModelField[TestedType, Model]{}
	field.Update(baseField)

	internalValue, toInternalValueErr := baseField.ToInternalValue(map[string]any{
		"some_items": passedValue,
	})

	assert.NoError(t, toInternalValueErr)
	assert.Equal(t, expectedValue, internalValue)
}

func TestSliceFieldToInternalValue(t *testing.T) {
	InternalValueCase[MockModel, string](t, []any{"bar", "baz"}, []string{"bar", "baz"})
	InternalValueCase[MockModel, float64](t, []any{1.0, 2.2}, []float64{1.0, 2.2})
	InternalValueCase[MockModel, map[string]string](
		t,
		[]any{map[string]string{"foo": "bar"}},
		[]map[string]string{{"foo": "bar"}},
	)
	InternalValueCase[MockModel, any](
		t,
		[]any{1.0, "foo", map[string]string{"foo": "bar"}},
		[]any{1.0, "foo", map[string]string{"foo": "bar"}},
	)
}

func TestSliceFieldNotACollection(t *testing.T) {
	var interf any
	for _, invalidTypeVar := range []any{
		interf,
		"foo",
		map[string]any{},
		map[string]string{},
		1,
		1.0,
		true,
	} {
		baseField := NewField[MockModel]("some_items")
		field := SliceModelField[string, MockModel]{}
		field.Update(baseField)

		_, toInternalValueErr := baseField.ToInternalValue(map[string]any{
			"some_items": invalidTypeVar,
		})

		assert.ErrorContains(t, toInternalValueErr, "Should be a collection")
	}
}

func TestSliceFieldOneOfCollectionItemsInvalidType(t *testing.T) {
	baseField := NewField[MockModel]("some_items")
	field := SliceModelField[float64, MockModel]{}
	field.Update(baseField)

	_, toInternalValueErr := baseField.ToInternalValue(map[string]any{
		"some_items": []any{1.0, 1.2, "foo"},
	})

	assert.ErrorContains(t, toInternalValueErr, "some_items[2] is not a valid float64")
}
