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

// DynamicGORM can be used to control the gorm.DB instance used by the query driver, for example
// in multi-tenant applications, where the database connection is determined by the request context.
func DynamicGORM[Model any](dbFunc func(*gin.Context) *gorm.DB) *gormdb.GormQueryDriver[Model] {
	return gormdb.Gorm[Model](
		gormdb.Dynamic(dbFunc),
	)
}
