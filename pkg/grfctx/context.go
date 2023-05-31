package grfctx

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"gorm.io/gorm"
)

type Context struct {
	Gin       *gin.Context
	db        *gorm.DB
	Container map[string]any
}

func (ctx *Context) Set(key string, value any) {
	ctx.Container[key] = value
}

func (ctx *Context) Get(key string) (any, bool) {
	val, ok := ctx.Container[key]
	return val, ok
}

func (ctx *Context) DB() *gorm.DB {
	return ctx.db.Session(&gorm.Session{NewDB: true})
}

func NewDBSession[Model any](ctx *Context) *gorm.DB {
	var m Model
	return ctx.db.Session(&gorm.Session{NewDB: true}).Model(&m)
}

func NewContext(ginCtx *gin.Context, db db.Resolver) (*Context, error) {
	theDb, dbErr := db.Resolve(ginCtx)
	if dbErr != nil {
		return nil, fmt.Errorf("failed to resolve db when creating request context: %w", dbErr)
	}
	return &Context{
		Gin:       ginCtx,
		db:        theDb,
		Container: make(map[string]any),
	}, nil
}

type ContextFactory struct {
	dbResolver db.Resolver
}

func (f *ContextFactory) New(ginCtx *gin.Context) (*Context, error) {
	return NewContext(ginCtx, f.dbResolver)
}

func NewContextFactory(db db.Resolver) *ContextFactory {
	return &ContextFactory{
		dbResolver: db,
	}
}
