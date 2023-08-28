package gormq

import (
	"fmt"

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

func CtxInitQuery(ctx *gin.Context) {
	ctx.Set("db:gorm:query", ORM(ctx))
}

func CtxSetQuery(ctx *gin.Context, db *gorm.DB) {
	ctx.Set("db:gorm:query", db)
}

func CtxQuery(ctx *gin.Context) *gorm.DB {
	return ctx.MustGet("db:gorm:query").(*gorm.DB)
}

func ORM(ctx *gin.Context) *gorm.DB {
	session := CtxGetGorm(ctx).Session(&gorm.Session{NewDB: true})
	session.Callback().Query().Before("gorm:query").Register("grf:before_query", func(d *gorm.DB) {
		fmt.Println(d.Statement.SQL.String())
	})
	session.Callback().Query().After("gorm:query").Register("grf:after_query", func(d *gorm.DB) {
		fmt.Println(d.Statement.SQL.String())
	})
	return session
}
