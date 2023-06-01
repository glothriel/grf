package grfctx

import (
	"github.com/gin-gonic/gin"
)

type Context struct {
	Gin *gin.Context
}

func (ctx *Context) Set(key string, value any) {
	ctx.Gin.Set(Prefix(key), value)
}

func (ctx *Context) Get(key string) (any, bool) {
	return ctx.Gin.Get(Prefix(key))
}

func New(ginCtx *gin.Context) (*Context, error) {
	return &Context{
		Gin: ginCtx,
	}, nil
}

func Prefix(key string) string {
	return "grfctx_" + key
}
