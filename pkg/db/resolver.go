package db

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Resolver interface {
	Resolve(*gin.Context) (*gorm.DB, error)
}

type StaticResolver struct {
	Db *gorm.DB
}

func (r *StaticResolver) Resolve(_ *gin.Context) (*gorm.DB, error) {
	return r.Db, nil
}

func NewStaticResolver(db *gorm.DB) Resolver {
	return &StaticResolver{
		Db: db,
	}
}
