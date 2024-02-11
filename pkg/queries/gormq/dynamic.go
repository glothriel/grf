package gormq

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type GormORMFactory interface {
	Create(*gin.Context) *gorm.DB
	InitQuery(*gorm.DB) *gorm.DB
}

type staticFactory struct {
	db *gorm.DB
}

func (s staticFactory) Create(ctx *gin.Context) *gorm.DB {
	return s.db
}

func (s staticFactory) InitQuery(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true})
}

func Static(db *gorm.DB) GormORMFactory {
	return staticFactory{db: db}
}

type dynamicFactory struct {
	dbFunc       func(*gin.Context) *gorm.DB
	copiedFields []string
}

func (d dynamicFactory) Create(ctx *gin.Context) *gorm.DB {
	return d.dbFunc(ctx)
}

func (d dynamicFactory) InitQuery(db *gorm.DB) *gorm.DB {
	newDb := db.Session(&gorm.Session{NewDB: true})
	for _, field := range d.copiedFields {
		if value, ok := db.InstanceGet(field); ok {
			newDb.InstanceSet(field, value)
		}
	}
	return newDb
}

func Dynamic(dbFunc func(*gin.Context) *gorm.DB, copiedCtxFields ...string) GormORMFactory {
	return dynamicFactory{dbFunc: dbFunc, copiedFields: copiedCtxFields}
}
