package gormdb

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockModel struct {
	ID  uint   `gorm:"primaryKey" json:"id"`
	Foo string `json:"foo"`
}

func prepareDb[Model any](t *testing.T) (*gin.Context, *GormDatabase[MockModel]) {
	gormDb, dbErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	migrateErr := gormDb.AutoMigrate(&MockModel{})
	database := Gorm[MockModel](gormDb)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.NoError(t, dbErr)
	assert.NoError(t, migrateErr)
	for _, middleware := range database.Middleware() {
		middleware(ctx)
	}
	return ctx, database
}

func TestGormDBInitializesQueryInMiddleware(t *testing.T) {
	// given
	ctx, _ := prepareDb[MockModel](t)

	// when
	results := []MockModel{}
	queryErr := CtxQuery[MockModel](ctx).Find(&results).Error

	// then
	assert.NoError(t, queryErr)
}

func TestGormDBCreateQuery(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	m, createErr := database.Queries().Create(ctx, &MockModel{Foo: "bar"})

	// then
	assert.NoError(t, createErr)
	assert.NotZero(t, m.ID)
}

func TestGormDBListQuery(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	_, createErr := database.Queries().Create(ctx, &MockModel{Foo: "bar"})
	list, listErr := database.Queries().List(ctx)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": uint(1), "foo": "bar"},
	}, list)
}

func TestGormDBRetrieveQuery(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	mockModel, createErr := database.Queries().Create(ctx, &MockModel{Foo: "bar"})
	item, retrieveErr := database.Queries().Retrieve(ctx, mockModel.ID)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, retrieveErr)
	assert.Equal(t, models.InternalValue{"id": uint(1), "foo": "bar"}, item)
}

func TestGormDBUpdateQuery(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	mockModel, createErr := database.Queries().Create(ctx, &MockModel{Foo: "bar"})
	item, updateErr := database.Queries().Update(ctx, &mockModel, &MockModel{ID: 1, Foo: "baz"}, mockModel.ID)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, updateErr)
	assert.Equal(t, MockModel{ID: 1, Foo: "baz"}, item)
}

func TestGormDBDeleteQuery(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	mockModel, createErr := database.Queries().Create(ctx, &MockModel{Foo: "bar"})
	deleteErr := database.Queries().Delete(ctx, mockModel, mockModel.ID)
	_, retrieveErr := database.Queries().Retrieve(ctx, mockModel.ID)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, deleteErr)
	assert.Error(t, retrieveErr)
}

func TestGormPagination(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	database.WithPagination(&NoPagination{}).Pagination().Apply(ctx)
	formatted, formatErr := database.Pagination().Format(ctx, []any{1, 2, 3})

	// then
	assert.NoError(t, formatErr)
	assert.Equal(t, []any{1, 2, 3}, formatted)
}

func TestGormOrder(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	for _, model := range []MockModel{
		{Foo: "bob"},
		{Foo: "alice"},
		{Foo: "david"},
	} {
		var localM = model
		_, createErr := database.Queries().Create(ctx, &localM)
		assert.NoError(t, createErr)
	}
	database.WithOrderBy("foo ASC")
	database.Order().Apply(ctx)
	list, listErr := database.Queries().List(ctx)

	// then
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": uint(2), "foo": "alice"},
		{"id": uint(1), "foo": "bob"},
		{"id": uint(3), "foo": "david"},
	}, list)
}

func TestGormFilter(t *testing.T) {
	// given
	ctx, database := prepareDb[MockModel](t)

	// when
	for _, model := range []MockModel{
		{Foo: "bob"},
		{Foo: "alice"},
		{Foo: "david"},
	} {
		var localM = model
		_, createErr := database.Queries().Create(ctx, &localM)
		assert.NoError(t, createErr)
	}
	database.WithFilter(func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		return db.Where("foo = ?", "alice")
	})
	database.Filter().Apply(ctx)
	list, listErr := database.Queries().List(ctx)

	// then
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": uint(2), "foo": "alice"},
	}, list)
}
