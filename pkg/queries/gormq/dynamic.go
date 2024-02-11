package gormq

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type GormORMFactory interface {
	Create(*gin.Context) *gorm.DB
}

type staticFactory struct {
	db *gorm.DB
}

func (s staticFactory) Create(*gin.Context) *gorm.DB {
	return s.db
}

func Static(db *gorm.DB) GormORMFactory {
	return staticFactory{db: db}
}

type dynamicFactory struct {
	dbFunc func(*gin.Context) *gorm.DB
}

func (d dynamicFactory) Create(ctx *gin.Context) *gorm.DB {
	return d.dbFunc(ctx)
}

func Dynamic(dbFunc func(*gin.Context) *gorm.DB) GormORMFactory {
	return dynamicFactory{dbFunc: dbFunc}
}
