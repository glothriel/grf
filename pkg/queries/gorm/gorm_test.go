package gorm

import (
	"net/http/httptest"
	"testing"
	"time"

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

func prepareGorm(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"))
	assert.NoError(t, err)
	sqlDb, sqlDbErr := db.DB()
	assert.NoError(t, sqlDbErr)
	sqlDb.SetMaxOpenConns(1)
	return db
}

func prepareCtx[Model any](t *testing.T, db ...*gorm.DB) (*gin.Context, *GormQueryDriver[Model]) {
	var theDB *gorm.DB
	if len(db) == 0 {
		theDB = prepareGorm(t)
	} else {
		theDB = db[0]
	}
	var m Model
	migrateErr := theDB.AutoMigrate(&m)
	queryDriver := Gorm[Model](theDB)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.NoError(t, migrateErr)
	for _, middleware := range queryDriver.Middleware() {
		middleware(ctx)
	}
	return ctx, queryDriver
}

func TestGormDBInitializesQueryInMiddleware(t *testing.T) {
	// given
	ctx, _ := prepareCtx[MockModel](t)

	// when
	results := []MockModel{}
	queryErr := CtxQuery(ctx).Find(&results).Error

	// then
	assert.NoError(t, queryErr)
}

func TestGormDBCreateQuery(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	created, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})

	// then
	assert.NoError(t, createErr)
	assert.NotZero(t, created["id"])
}

func TestGormDBCreateQueryNotAValidModel(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	_, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": time.Duration(1)})

	// then
	assert.Error(t, createErr)
}

func TestGormDBListQuery(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	_, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	list, listErr := queryDriver.CRUD().List(ctx)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": uint(1), "foo": "bar"},
	}, list)
}

func TestGormDBRetrieveQuery(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	intVal, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	item, retrieveErr := queryDriver.CRUD().Retrieve(ctx, intVal["id"])

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, retrieveErr)
	assert.Equal(t, models.InternalValue{"id": uint(1), "foo": "bar"}, item)
}

func TestGormDBUpdateQueryDoesNotExists(t *testing.T) {
	t.Skip("This has to be fixed, currently fails, consider debugging query executed by gorm")
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	_, updateErr := queryDriver.CRUD().Update(ctx, models.InternalValue{}, models.InternalValue{}, uint(1))

	// then
	assert.Error(t, updateErr)
}

func TestGormDBUpdateQueryNotAValidModel(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	intVal, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	_, updateErr := queryDriver.CRUD().Update(ctx, intVal, models.InternalValue{
		"foo": time.Duration(1),
	}, intVal["id"])

	// then
	assert.NoError(t, createErr)
	assert.Error(t, updateErr)
}

func TestGormDBUpdateQuery(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	intVal, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	item, updateErr := queryDriver.CRUD().Update(
		ctx, intVal, models.InternalValue{"id": intVal["id"], "foo": "baz"}, intVal["id"],
	)

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, updateErr)
	assert.Equal(t, models.InternalValue{"foo": "baz", "id": intVal["id"]}, item)
}

func TestGormDBDestroyQuery(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	intVal, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	deleteErr := queryDriver.CRUD().Destroy(ctx, intVal["id"])
	_, retrieveErr := queryDriver.CRUD().Retrieve(ctx, intVal["id"])

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, deleteErr)
	assert.Error(t, retrieveErr)
}

func TestGormDBDestroyQueryDoesNotExist(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	deleteErr := queryDriver.CRUD().Destroy(ctx, uint(1))

	// then
	assert.Error(t, deleteErr)
}

func TestGormPagination(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	queryDriver.WithPagination(&NoPagination{}).Pagination().Apply(ctx)
	formatted, formatErr := queryDriver.Pagination().Format(ctx, []any{1, 2, 3})

	// then
	assert.NoError(t, formatErr)
	assert.Equal(t, []any{1, 2, 3}, formatted)
}

func TestGormOrder(t *testing.T) {
	// given
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	for _, model := range []models.InternalValue{
		{"foo": "bob"},
		{"foo": "alice"},
		{"foo": "david"},
	} {
		var localM = model
		_, createErr := queryDriver.CRUD().Create(ctx, localM)
		assert.NoError(t, createErr)
	}
	queryDriver.WithOrderBy("foo ASC")
	queryDriver.Order().Apply(ctx)
	list, listErr := queryDriver.CRUD().List(ctx)

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
	ctx, queryDriver := prepareCtx[MockModel](t)

	// when
	for _, model := range []models.InternalValue{
		{"foo": "bob"},
		{"foo": "alice"},
		{"foo": "david"},
	} {
		var localM = model
		_, createErr := queryDriver.CRUD().Create(ctx, localM)
		assert.NoError(t, createErr)
	}
	queryDriver.WithFilter(func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		return db.Where("foo = ?", "alice")
	})
	queryDriver.Filter().Apply(ctx)
	list, listErr := queryDriver.CRUD().List(ctx)

	// then
	assert.NoError(t, listErr)
	assert.Equal(t, []models.InternalValue{
		{"id": uint(2), "foo": "alice"},
	}, list)
}
