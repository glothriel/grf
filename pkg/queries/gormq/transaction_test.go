package gormq

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type Log struct {
	ID      int64  `gorm:"primaryKey" json:"id"`
	Content string `gorm:"type:varchar(255)" json:"content"`
}

func TestCreateTx(t *testing.T) {
	// given
	db := prepareGorm(t)
	ctx, queryDriver := prepareCtx[MockModel](t, db)
	logsCtx, logQD := prepareCtx[Log](t, db)

	// when
	created, createErr := queryDriver.CRUD().WithCreate(CreateTx(BeforeCreate(
		func(ctx *gin.Context, iv models.InternalValue, db *gorm.DB) (models.InternalValue, error) {
			log := Log{Content: "before"}
			return iv, db.Model(&Log{}).Create(&log).Error
		},
	), AfterCreate(
		func(ctx *gin.Context, iv models.InternalValue, db *gorm.DB) (models.InternalValue, error) {
			log := Log{Content: "after"}
			return iv, db.Model(&Log{}).Create(&log).Error
		},
	))(queryDriver.CRUD().Create)).Create(ctx, models.InternalValue{"foo": "bar"})

	logs, logsListErr := logQD.CRUD().List(logsCtx)
	retrieved, retrieveErr := queryDriver.CRUD().Retrieve(ctx, created["id"])

	// then
	assert.NoError(t, createErr)
	assert.NoError(t, logsListErr)
	assert.NoError(t, retrieveErr)
	assert.Equal(t, []models.InternalValue{
		{
			"id":      int64(1),
			"content": "before",
		},
		{
			"id":      int64(2),
			"content": "after",
		},
	}, logs)
	assert.Equal(t, models.InternalValue{
		"id":  uint(1),
		"foo": "bar",
	}, retrieved)
}

func TestCreateTxErr(t *testing.T) {
	tests := []struct {
		name  string
		hooks CreateTxHooks
	}{
		{
			name: "BeforeCreate error",
			hooks: AfterCreate(
				func(ctx *gin.Context, iv models.InternalValue, db *gorm.DB) (models.InternalValue, error) {
					log := Log{Content: "after"}
					db.Model(&Log{}).Create(&log)
					return iv, errors.New("foobar")
				},
			),
		},
		{
			name: "BeforeCreate error",
			hooks: BeforeCreate(
				func(ctx *gin.Context, iv models.InternalValue, db *gorm.DB) (models.InternalValue, error) {
					log := Log{Content: "before"}
					db.Model(&Log{}).Create(&log)
					return iv, errors.New("foobar")
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			db := prepareGorm(t)
			ctx, queryDriver := prepareCtx[MockModel](t, db)
			logsCtx, logQD := prepareCtx[Log](t, db)

			// when
			created, createErr := queryDriver.CRUD().WithCreate(
				CreateTx(tt.hooks)(queryDriver.CRUD().Create),
			).Create(ctx, models.InternalValue{"foo": "bar"})

			logList, logsListErr := logQD.CRUD().List(logsCtx)
			_, retrieveErr := queryDriver.CRUD().Retrieve(ctx, created["id"])

			// then
			assert.NoError(t, logsListErr)
			assert.Len(t, logList, 0)

			assert.Error(t, createErr)
			assert.Error(t, retrieveErr)
		})
	}
}

func TestUpdateTx(t *testing.T) {
	// given
	db := prepareGorm(t)
	ctx, queryDriver := prepareCtx[MockModel](t, db)
	logsCtx, logQD := prepareCtx[Log](t, db)
	prepareCtx[Log](t, db)

	created, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	assert.NoError(t, createErr)

	// when
	_, updateErr := queryDriver.CRUD().WithUpdate(UpdateTx(BeforeUpdate(
		func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any, db *gorm.DB) (models.InternalValue, error) {
			log := Log{Content: "before"}
			return new, db.Model(&Log{}).Create(&log).Error
		},
	), AfterUpdate(
		func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any, db *gorm.DB) (models.InternalValue, error) {
			log := Log{Content: "after"}
			return new, db.Model(&Log{}).Create(&log).Error
		},
	))(queryDriver.CRUD().Update)).Update(ctx,
		models.InternalValue{"id": created["id"], "foo": "bar"},
		models.InternalValue{"id": created["id"], "foo": "baz"},
		created["id"],
	)

	logs, logsListErr := logQD.CRUD().List(logsCtx)
	retrieved, retrieveErr := queryDriver.CRUD().Retrieve(ctx, created["id"])

	// then
	assert.NoError(t, updateErr)
	assert.NoError(t, logsListErr)
	assert.NoError(t, retrieveErr)
	assert.Equal(t, []models.InternalValue{
		{
			"id":      int64(1),
			"content": "before",
		},
		{
			"id":      int64(2),
			"content": "after",
		},
	}, logs)
	assert.Equal(t, models.InternalValue{
		"id":  uint(1),
		"foo": "baz",
	}, retrieved)
}

func TestUpdateTxErr(t *testing.T) {
	tests := []struct {
		name  string
		hooks UpdateTxHooks
	}{
		{
			name: "BeforeUpdate error",
			hooks: AfterUpdate(
				func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any, db *gorm.DB) (models.InternalValue, error) {
					log := Log{Content: "after"}
					db.Model(&Log{}).Create(&log)
					return new, errors.New("foobar")
				},
			),
		},
		{
			name: "BeforeUpdate error",
			hooks: BeforeUpdate(
				func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any, db *gorm.DB) (models.InternalValue, error) {
					log := Log{Content: "before"}
					db.Model(&Log{}).Create(&log)
					return new, errors.New("foobar")
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			db := prepareGorm(t)
			ctx, queryDriver := prepareCtx[MockModel](t, db)
			logsCtx, logQD := prepareCtx[Log](t, db)

			created, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
			assert.NoError(t, createErr)

			// when
			_, updateErr := queryDriver.CRUD().WithUpdate(
				UpdateTx(tt.hooks)(queryDriver.CRUD().Update),
			).Update(ctx,
				models.InternalValue{"id": created["id"], "foo": "bar"},
				models.InternalValue{"id": created["id"], "foo": "baz"},
				created["id"],
			)

			logList, logsListErr := logQD.CRUD().List(logsCtx)

			// then
			assert.NoError(t, logsListErr)
			assert.Len(t, logList, 0)

			assert.Error(t, updateErr)
		})
	}
}

func TestDestroyTx(t *testing.T) {
	// given
	db := prepareGorm(t)
	ctx, queryDriver := prepareCtx[MockModel](t, db)
	logsCtx, logQD := prepareCtx[Log](t, db)
	prepareCtx[Log](t, db)

	created, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
	assert.NoError(t, createErr)

	// when
	destroyErr := queryDriver.CRUD().WithDestroy(DestroyTx(BeforeDestroy(
		func(ctx *gin.Context, id any, db *gorm.DB) error {
			log := Log{Content: "before"}
			return db.Model(&Log{}).Create(&log).Error
		},
	), AfterDestroy(
		func(ctx *gin.Context, id any, db *gorm.DB) error {
			log := Log{Content: "after"}
			return db.Model(&Log{}).Create(&log).Error
		},
	))(queryDriver.CRUD().Destroy)).Destroy(ctx, created["id"])

	logs, logsListErr := logQD.CRUD().List(logsCtx)
	_, retrieveErr := queryDriver.CRUD().Retrieve(ctx, created["id"])

	// then
	assert.NoError(t, destroyErr)
	assert.NoError(t, logsListErr)
	assert.Error(t, retrieveErr)
	assert.Equal(t, []models.InternalValue{
		{
			"id":      int64(1),
			"content": "before",
		},
		{
			"id":      int64(2),
			"content": "after",
		},
	}, logs)
}

func TestDestroyTxErr(t *testing.T) {
	tests := []struct {
		name  string
		hooks DestroyTxHooks
	}{
		{
			name: "BeforeDestroy error",
			hooks: AfterDestroy(
				func(ctx *gin.Context, id any, db *gorm.DB) error {
					log := Log{Content: "after"}
					db.Model(&Log{}).Create(&log)
					return errors.New("foobar")
				},
			),
		},
		{
			name: "BeforeDestroy error",
			hooks: BeforeDestroy(
				func(ctx *gin.Context, id any, db *gorm.DB) error {
					log := Log{Content: "before"}
					db.Model(&Log{}).Create(&log)
					return errors.New("foobar")
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			db := prepareGorm(t)
			ctx, queryDriver := prepareCtx[MockModel](t, db)
			logsCtx, logQD := prepareCtx[Log](t, db)

			created, createErr := queryDriver.CRUD().Create(ctx, models.InternalValue{"foo": "bar"})
			assert.NoError(t, createErr)

			// when
			destroyErr := queryDriver.CRUD().WithDestroy(
				DestroyTx(tt.hooks)(queryDriver.CRUD().Destroy),
			).Destroy(ctx, created["id"])
			logList, logsListErr := logQD.CRUD().List(logsCtx)
			mockList, mockListErr := queryDriver.CRUD().List(ctx)

			// then
			assert.NoError(t, logsListErr)
			assert.NoError(t, mockListErr)
			assert.Len(t, logList, 0)
			assert.Len(t, mockList, 1)
			assert.Error(t, destroyErr)
		})
	}
}
