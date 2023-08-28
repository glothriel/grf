package fields

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
)

type fieldFuncsMocks struct {
	toRepresentationArgInternalValue models.InternalValue
	toRepresentationArgName          string
	toRepresentationArgCtx           *gin.Context

	toInternalValueArgReprModel map[string]any
	toInternalValueArgName      string
	toInternalValueArgCtx       *gin.Context
}

func (m *fieldFuncsMocks) toRepresentationMock(intVal models.InternalValue, name string, ctx *gin.Context) (any, error) {
	m.toRepresentationArgInternalValue = intVal
	m.toRepresentationArgName = name
	m.toRepresentationArgCtx = ctx

	return nil, nil
}

func (m *fieldFuncsMocks) toInternalValueMock(reprModel map[string]any, name string, ctx *gin.Context) (any, error) {
	m.toInternalValueArgReprModel = reprModel
	m.toInternalValueArgName = name
	m.toInternalValueArgCtx = ctx

	return nil, nil
}

func TestFieldName(t *testing.T) {
	// given
	field := Field[struct{}]{
		name: "test",
	}

	// when
	name := field.Name()

	// then
	assert.Equal(t, "test", name)
}

func TestFieldToRepresentation(t *testing.T) {
	// given
	mocks := &fieldFuncsMocks{}
	field := Field[struct{}]{
		name:               "foo",
		representationFunc: mocks.toRepresentationMock,
	}
	intVal := models.InternalValue{"foo": "bar"}
	ginCtx := &gin.Context{}

	// when
	field.ToRepresentation(intVal, ginCtx)

	// then
	assert.Equal(t, intVal, mocks.toRepresentationArgInternalValue)
	assert.Equal(t, "foo", mocks.toRepresentationArgName)
	assert.Equal(t, ginCtx, mocks.toRepresentationArgCtx)
}

func TestFieldToInternalValue(t *testing.T) {
	// given
	mocks := &fieldFuncsMocks{}
	field := Field[struct{}]{
		name:              "foo",
		internalValueFunc: mocks.toInternalValueMock,
	}
	reprModel := map[string]any{"foo": "bar"}
	ginCtx := &gin.Context{}

	// when
	field.ToInternalValue(reprModel, ginCtx)

	// then
	assert.Equal(t, reprModel, mocks.toInternalValueArgReprModel)
	assert.Equal(t, "foo", mocks.toInternalValueArgName)
	assert.Equal(t, ginCtx, mocks.toInternalValueArgCtx)
}

func TestFieldReadOnly(t *testing.T) {
	// given
	field := Field[struct{}]{}

	// when
	field.ReadOnly()

	// then
	assert.True(t, field.Readable)
	assert.False(t, field.Writable)
}

func TestFieldWriteOnly(t *testing.T) {
	// given
	field := Field[struct{}]{}

	// when
	field.WriteOnly()

	// then
	assert.False(t, field.Readable)
	assert.True(t, field.Writable)
}

func TestFieldReadWrite(t *testing.T) {
	// given
	field := Field[struct{}]{}

	// when
	field.ReadWrite()

	// then
	assert.True(t, field.Readable)
	assert.True(t, field.Writable)
}

func TestFieldWithRepresentationFunc(t *testing.T) {
	// given
	field := Field[struct{}]{}
	funcCalls := 0
	f := func(models.InternalValue, string, *gin.Context) (any, error) {
		funcCalls++
		return nil, nil
	}

	// when
	field.WithRepresentationFunc(f)
	field.ToRepresentation(nil, nil)

	// then
	assert.Equal(t, 1, funcCalls)
}

func TestFieldWithInternalValueFunc(t *testing.T) {
	// given
	field := Field[struct{}]{}
	funcCalls := 0
	f := func(map[string]any, string, *gin.Context) (any, error) {
		funcCalls++
		return nil, nil
	}

	// when
	field.WithInternalValueFunc(f)
	field.ToInternalValue(nil, nil)

	// then
	assert.Equal(t, 1, funcCalls)
}

type mockModel struct {
	Field string `json:"field"`
}

func TestFieldDefaultFuncs(t *testing.T) {
	// given
	field := NewField[mockModel]("field")

	// when
	intVal, intValErr := field.ToInternalValue(map[string]any{"field": "foo"}, nil)
	reprVal, reprValErr := field.ToRepresentation(models.InternalValue{"field": "foo"}, nil)

	// then
	assert.Equal(t, "foo", intVal)
	assert.Nil(t, intValErr)
	assert.Equal(t, "foo", reprVal)
	assert.Nil(t, reprValErr)
}
