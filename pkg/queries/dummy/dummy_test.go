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
		{"id": uint(1), "foo": "bar"},
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
	assert.Equal(t, models.InternalValue{"id": uint(1), "foo": "bar"}, retrieved)
}

func TestDummyRetrieveDNE(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	_, retrieveErr := driver.CRUD().Retrieve(ctx, 2)

	// then
	assert.Equal(t, common.ErrorNotFound, retrieveErr)
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

func TestDummyUpdateDNE(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	_, updateErr := driver.CRUD().Update(ctx, models.InternalValue{
		"foo": "bar",
	}, models.InternalValue{
		"foo": "baz",
	}, 2)

	// then
	assert.Equal(t, common.ErrorNotFound, updateErr)
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
		"foo": "baz", "id": uint(2),
	}, created)
}

func TestDummyDestroy(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	deleteErr := driver.CRUD().Destroy(ctx, 1)
	_, retrieveErr := driver.CRUD().Retrieve(ctx, 1)

	// then
	assert.NoError(t, deleteErr)
	assert.Equal(t, common.ErrorNotFound, retrieveErr)
}

func TestDummyDestroyDNE(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	deleteErr := driver.CRUD().Destroy(ctx, 2)

	// then
	assert.Equal(t, common.ErrorNotFound, deleteErr)
}

func TestIDGeneratorInt(t *testing.T) {
	// given
	generator := newIDGenerator[struct {
		ID int `json:"id"`
	}](map[any]models.InternalValue{})

	// when
	id := generator()

	// then
	assert.Equal(t, 1, id)
}

func TestIDGeneratorString(t *testing.T) {
	// given
	generator := newIDGenerator[struct {
		ID string `json:"id"`
	}](map[any]models.InternalValue{})

	// when
	id := generator()

	// then
	assert.Regexp(t, "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", id)
}

func TestMiddleware(t *testing.T) {
	// given
	driver := InMemoryDriver(MockModel{Foo: "bar"})

	// when
	m := driver.Middleware()

	// then
	assert.Len(t, m, 0)
}

func TestFormat(t *testing.T) {
	// given
	pagination := dummyPagination[struct{}]{}

	// when
	formatted, err := pagination.Format(nil, []any{1})

	// then
	assert.Equal(t, []any{1}, formatted)
	assert.NoError(t, err)
}
