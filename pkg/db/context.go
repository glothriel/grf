package db

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func CtxGetGorm(ctx *gin.Context) *gorm.DB {
	anyVal, ok := ctx.Get("gorm")
	if !ok {
		logrus.Fatal("gorm not found in the context. Was it initialized earlier?")
	}

	theVal, ok := anyVal.(*gorm.DB)
	if !ok {
		logrus.Fatal("gorm has wrong type in the context. Was it initialized earlier?")
	}
	return theVal
}

func CtxSetGorm(dbResolver Resolver) func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		db, err := dbResolver.Resolve(ctx)
		if err != nil {
			logrus.Fatal("Failed to resolve db")
		}

		logrus.Error("SET GORM")
		ctx.Set("gorm", db)
	}
}

func ORM[Model any](ctx *gin.Context) *gorm.DB {
	var entity Model
	return CtxGetGorm(ctx).Session(&gorm.Session{NewDB: true}).Model(&entity)
}
