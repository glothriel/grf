package common

import "github.com/gin-gonic/gin"

type QueryMod interface {
	Apply(*gin.Context)
}

type Pagination interface {
	QueryMod
	Format(*gin.Context, []any) (any, error)
}

type CompositeQueryMod struct {
	children []QueryMod
}

func (c CompositeQueryMod) Apply(ctx *gin.Context) {
	for _, child := range c.children {
		child.Apply(ctx)
	}
}

func NewCompositeQueryMod(children ...QueryMod) CompositeQueryMod {
	return CompositeQueryMod{children: children}
}
