package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func FromRepresentationSuccessTestCase[T any](t *testing.T, value []any, expected []T) {
	var s SliceModelField[T]
	err := s.FromRepresentation(value)
	assert.NoError(t, err)
	for i, item := range expected {
		assert.Equal(t, item, s[i])
	}
	assert.Len(t, s, len(expected))
}

func FromRepresentationErrorTestCase[T any](t *testing.T, value any) {
	var s SliceModelField[T]
	err := s.FromRepresentation(value)
	assert.Error(t, err)
}

func TestSliceModelFieldFromRepresentationNoError(t *testing.T) {
	FromRepresentationSuccessTestCase(t, []any{1, 2, 3}, []int{1, 2, 3})
	FromRepresentationSuccessTestCase(t, []any{"1", "2", "3"}, []string{"1", "2", "3"})
	FromRepresentationSuccessTestCase(t, []any{true, false, true}, []bool{true, false, true})
	FromRepresentationSuccessTestCase(t, []any{1.1, 2.2, 3.3}, []float64{1.1, 2.2, 3.3})

}

func TestSliceModelFieldFromRepresentationError(t *testing.T) {
	FromRepresentationErrorTestCase[int](t, []any{1, 2, "3"})
	FromRepresentationErrorTestCase[string](t, []any{1, 2, 3})
	FromRepresentationErrorTestCase[bool](t, []any{1, 2, 3})
	FromRepresentationErrorTestCase[float64](t, []any{1, 2, 3})
	FromRepresentationErrorTestCase[float64](t, 1.337)
}

func TestSliceModelFieldToRepresentation(t *testing.T) {
	// given
	var s SliceModelField[int]
	s = []int{1, 2, 3}

	// when
	value, err := s.ToRepresentation()

	// then
	assert.NoError(t, err)
	assert.Equal(t, SliceModelField[int]{1, 2, 3}, value)
}

func TestSliceModelFieldScan(t *testing.T) {
	// given
	var s SliceModelField[int]

	// when
	err := s.Scan([]byte(`[1,2,3]`))

	// then
	assert.NoError(t, err)
}

func TestSliceModelFieldScanErrorJSON(t *testing.T) {
	// given
	var s SliceModelField[int]

	// when
	err := s.Scan([]byte(`[1,2,3`))

	// then
	assert.Error(t, err)
}

func TestSliceModelFieldScanErrorNotBytes(t *testing.T) {
	// given
	var s SliceModelField[int]

	// when
	err := s.Scan(1.337)

	// then
	assert.Error(t, err)
}

func TestSliceModelFieldValue(t *testing.T) {
	// given
	s := SliceModelField[int]{1, 2, 3}

	// when
	value, err := s.Value()

	// then
	assert.NoError(t, err)
	assert.Equal(t, []byte(`[1,2,3]`), value)
}

func TestSliceModelFieldValueEmpty(t *testing.T) {
	// given
	s := SliceModelField[int]{}

	// when
	value, err := s.Value()

	// then
	assert.NoError(t, err)
	assert.Equal(t, []byte(`[]`), value)
}
