package gormq

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func CtxSetFactory(ctx *gin.Context, factory GormORMFactory) {
	ctx.Set("db:gorm:factory", factory)
}

func CtxGetFactory(ctx *gin.Context) GormORMFactory {
	anyVal, ok := ctx.Get("db:gorm:factory")
	if !ok {
		logrus.Panic("gorm factory not found in the context. Was it initialized earlier?")
	}

	theVal, ok := anyVal.(GormORMFactory)
	if !ok {
		logrus.Panic("gorm factory has wrong type in the context. Was it initialized earlier?")
	}
	return theVal
}

func CtxInitQuery(ctx *gin.Context) {
	ctx.Set("db:gorm:query", New(ctx))
}

func CtxSetQuery(ctx *gin.Context, db *gorm.DB) {
	ctx.Set("db:gorm:query", db)
}

func CtxQuery(ctx *gin.Context) *gorm.DB {
	return ctx.MustGet("db:gorm:query").(*gorm.DB)
}

func New(ctx *gin.Context) *gorm.DB {
	return CtxGetFactory(ctx).Create(ctx)
}
