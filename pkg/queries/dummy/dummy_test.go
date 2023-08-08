package dummy

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/stretchr/testify/assert"
)

type MockModel struct {
	ID  uint   `json:"id"`
	Foo string `json:"foo"`
}

func TestDummyList(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	driver.Pagination().Apply(ctx)
	driver.Filter().Apply(ctx)
	driver.Order().Apply(ctx)
	list, listErr := driver.CRUD().List(ctx)

	// then
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": 1, "foo": "bar"},
	}, list)
}

func TestDummyRetrievie(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	retrieved, retrieveErr := driver.CRUD().Retrieve(ctx, 1)

	// then
	assert.NoError(t, retrieveErr)
	assert.Equal(t, models.InternalValue{"id": 1, "foo": "bar"}, retrieved)
}

func TestDummyUpdate(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	updated, updateErr := driver.CRUD().Update(ctx, models.InternalValue{
		"foo": "bar",
	}, models.InternalValue{
		"foo": "baz",
	}, 1)

	// then
	assert.NoError(t, updateErr)
	assert.Equal(t, models.InternalValue{
		"foo": "baz",
	}, updated)
}

func TestDummyCreate(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	created, createErr := driver.CRUD().Create(ctx, models.InternalValue{
		"foo": "baz",
	})

	// then
	assert.NoError(t, createErr)
	assert.Equal(t, models.InternalValue{
		"foo": "baz", "id": 2,
	}, created)
}

func TestDummyDelete(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	deleteErr := driver.CRUD().Delete(ctx, 1)
	_, retrieveErr := driver.CRUD().Retrieve(ctx, 1)

	// then
	assert.NoError(t, deleteErr)
	assert.Equal(t, common.ErrorNotFound, retrieveErr)
}
