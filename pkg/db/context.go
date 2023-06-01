package db

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func CtxGetGorm(ctx *grfctx.Context) *gorm.DB {
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

func CtxSetGorm(dbResolver Resolver) func(ctx *grfctx.Context) {
	return func(ctx *grfctx.Context) {
		db, err := dbResolver.Resolve(ctx)
		if err != nil {
			logrus.Fatal("Failed to resolve db")
		}
		ctx.Set("gorm", db)
	}
}

func CtxNewQuery[Model any](ctx *grfctx.Context) *gorm.DB {
	var entity Model
	return CtxGetGorm(ctx).Session(&gorm.Session{NewDB: true}).Model(&entity)
}
