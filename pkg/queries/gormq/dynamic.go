package gormq

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type GormORMFactory interface {
	Create(*gin.Context) *gorm.DB
}

type dynamicFactory struct {
	dbFunc func(*gin.Context) *gorm.DB
}

func (d dynamicFactory) Create(ctx *gin.Context) *gorm.DB {
	return d.dbFunc(ctx)
}

// Static creates a GormORMFactory that always returns the same gorm.DB instance,
// creating new sessions for each query.
func Static(db *gorm.DB) GormORMFactory {
	return dynamicFactory{
		dbFunc: func(*gin.Context) *gorm.DB {
			return db.Session(&gorm.Session{NewDB: true})
		},
	}
}

// Dynamic creates a GormORMFactory that provides a gorm.DB instance for each request,
// using the provided function. Make sure to create a new session for each query (if needed).
func Dynamic(dbFunc func(*gin.Context) *gorm.DB) GormORMFactory {
	return dynamicFactory{dbFunc: dbFunc}
}
