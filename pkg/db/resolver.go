package db

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"gorm.io/gorm"
)

type Resolver interface {
	Resolve(*grfctx.Context) (*gorm.DB, error)
}

type StaticResolver struct {
	Db *gorm.DB
}

func (r *StaticResolver) Resolve(_ *grfctx.Context) (*gorm.DB, error) {
	return r.Db, nil
}

func NewStaticResolver(db *gorm.DB) Resolver {
	return &StaticResolver{
		Db: db,
	}
}
