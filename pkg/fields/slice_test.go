package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockModel struct{}

func InternalValueCase[Model any, TestedType any](t *testing.T, passedValue any, expectedValue any) {
	baseField := NewField[Model]("some_items")
	field := SliceField[TestedType, Model]{}
	field.Update(baseField)

	internalValue, toInternalValueErr := baseField.ToInternalValue(map[string]interface{}{
		"some_items": passedValue,
	})

	assert.NoError(t, toInternalValueErr)
	assert.Equal(t, expectedValue, internalValue)
}

func TestSliceFieldToInternalValue(t *testing.T) {
	InternalValueCase[MockModel, string](t, []interface{}{"bar", "baz"}, []string{"bar", "baz"})
	InternalValueCase[MockModel, float64](t, []interface{}{1.0, 2.2}, []float64{1.0, 2.2})
	InternalValueCase[MockModel, map[string]string](
		t,
		[]interface{}{map[string]string{"foo": "bar"}},
		[]map[string]string{{"foo": "bar"}},
	)
}

func TestSliceFieldNotACollection(t *testing.T) {
	var interf interface{}
	for _, invalidTypeVar := range []interface{}{
		interf,
		"foo",
		map[string]any{},
		map[string]string{},
		1,
		1.0,
		true,
	} {
		baseField := NewField[MockModel]("some_items")
		field := SliceField[string, MockModel]{}
		field.Update(baseField)

		_, toInternalValueErr := baseField.ToInternalValue(map[string]interface{}{
			"some_items": invalidTypeVar,
		})

		assert.ErrorContains(t, toInternalValueErr, "Should be a collection")
	}
}

func TestSliceFieldOneOfCollectionItemsInvalidType(t *testing.T) {
	baseField := NewField[MockModel]("some_items")
	field := SliceField[float64, MockModel]{}
	field.Update(baseField)

	_, toInternalValueErr := baseField.ToInternalValue(map[string]interface{}{
		"some_items": []any{1.0, 1.2, "foo"},
	})

	assert.ErrorContains(t, toInternalValueErr, "some_items[2] is not a valid float64")
}
