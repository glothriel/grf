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
	dbFunc func(*gin.Context) *gorm.DB
}

func (d dynamicFactory) Create(ctx *gin.Context) *gorm.DB {
	return d.dbFunc(ctx)
}

func (d dynamicFactory) InitQuery(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true})
}

func Dynamic(dbFunc func(*gin.Context) *gorm.DB) GormORMFactory {
	return dynamicFactory{dbFunc: dbFunc}
}

type contextFieldsCopyingFactory struct {
	child  GormORMFactory
	fields []string
}

func (c contextFieldsCopyingFactory) Create(ctx *gin.Context) *gorm.DB {
	return c.child.Create(ctx)
}

func (c contextFieldsCopyingFactory) InitQuery(db *gorm.DB) *gorm.DB {
	newDb := c.child.InitQuery(db)

	for _, field := range c.fields {
		if value, ok := db.InstanceGet(field); ok {
			newDb.InstanceSet(field, value)
		}
	}
	return newDb
}

func ContextFieldsCopying(child GormORMFactory, fields ...string) GormORMFactory {
	return contextFieldsCopyingFactory{child: child, fields: fields}
}
