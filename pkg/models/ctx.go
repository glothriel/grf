package models

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/sirupsen/logrus"
)

func CtxStoreInternalValue(ctx *grfctx.Context, intVal map[string]any) {
	ctx.Set("internal_value", intVal)
}

func CtxGetInternalValue(ctx *grfctx.Context) InternalValue {
	anyVal, ok := ctx.Get("internal_value")
	if !ok {
		logrus.Fatal("InternalValue not found in the context. Was it initialized earlier?")
	}

	theVal, ok := anyVal.(InternalValue)
	if !ok {
		logrus.Fatal("InternalValue has wrong type in the context. Was it initialized earlier?")
	}
	return theVal
}
