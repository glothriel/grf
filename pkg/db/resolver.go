package db

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Resolver interface {
	Resolve(*gin.Context) (*gorm.DB, error)
}

type DefaultResolver struct {
	Db *gorm.DB
}

func (r *DefaultResolver) Resolve(_ *gin.Context) (*gorm.DB, error) {
	return r.Db, nil
}
