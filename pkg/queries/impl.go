package queries

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries/dummy"
	gormdb "github.com/glothriel/grf/pkg/queries/gormq"
	"gorm.io/gorm"
)

func InMemory[Model any](seed ...Model) *dummy.InMemoryQueryDriver[Model] {
	return dummy.InMemoryDriver(seed...)
}

func GORM[Model any](db *gorm.DB) *gormdb.GormQueryDriver[Model] {
	return gormdb.Gorm[Model](gormdb.Static(db))
}

func DynamicGORM[Model any](dbFunc func(*gin.Context) *gorm.DB) *gormdb.GormQueryDriver[Model] {
	return gormdb.Gorm[Model](gormdb.Dynamic(dbFunc))
}
