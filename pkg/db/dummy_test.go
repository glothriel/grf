package db

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
)

type MockModel struct {
	Foo string `json:"foo"`
}

func TestDummyList(t *testing.T) {
	// given
	database := Dummy(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	database.Pagination().Apply(ctx)
	database.Filter().Apply(ctx)
	database.Order().Apply(ctx)
	list, listErr := database.Queries().List(ctx)

	// then
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"foo": "bar"},
	}, list)
}

func TestDummyRetrievie(t *testing.T) {
	// given
	database := Dummy(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	retrieved, retrieveErr := database.Queries().Retrieve(ctx, 1)

	// then
	assert.NoError(t, retrieveErr)
	assert.Equal(t, models.InternalValue{"foo": "bar"}, retrieved)
}

func TestDummyUpdate(t *testing.T) {
	// given
	database := Dummy(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	updated, updateErr := database.Queries().Update(ctx, &MockModel{Foo: "baz"}, &MockModel{Foo: "baz"}, 1)

	// then
	assert.NoError(t, updateErr)
	assert.Equal(t, MockModel{Foo: "bar"}, updated) // Dummy does not update
}

func TestDummyCreate(t *testing.T) {
	// given
	database := Dummy(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	created, createErr := database.Queries().Create(ctx, &MockModel{Foo: "baz"})

	// then
	assert.NoError(t, createErr)
	assert.Equal(t, MockModel{Foo: "bar"}, created) // Dummy does not create
}

func TestDummyDelete(t *testing.T) {
	// given
	database := Dummy(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	deleteErr := database.Queries().Delete(ctx, MockModel{}, 1)

	// then
	assert.NoError(t, deleteErr)
}
