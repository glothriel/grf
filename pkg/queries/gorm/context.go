package gorm

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func CtxGetGorm(ctx *gin.Context) *gorm.DB {
	anyVal, ok := ctx.Get("db:gorm")
	if !ok {
		logrus.Panic("gorm not found in the context. Was it initialized earlier?")
	}

	theVal, ok := anyVal.(*gorm.DB)
	if !ok {
		logrus.Panic("gorm has wrong type in the context. Was it initialized earlier?")
	}
	return theVal
}

func CtxSetGorm(ctx *gin.Context, gormDB *gorm.DB) {
	ctx.Set("db:gorm", gormDB)
}

func CtxInitQuery[Model any](ctx *gin.Context) {
	var m Model
	ctx.Set("db:gorm:query", ORM[Model](ctx).Model(&m))
}

func CtxSetQuery[Model any](ctx *gin.Context, db *gorm.DB) {
	ctx.Set("db:gorm:query", db)
}

func CtxQuery[Model any](ctx *gin.Context) *gorm.DB {
	return ctx.MustGet("db:gorm:query").(*gorm.DB)
}

func ORM[Model any](ctx *gin.Context) *gorm.DB {
	var entity Model
	return CtxGetGorm(ctx).Session(&gorm.Session{NewDB: true}).Model(&entity)
}
