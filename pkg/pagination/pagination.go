package pagination

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Pagination interface {
	Apply(*gin.Context, *gorm.DB) *gorm.DB
	Format(*gin.Context, []any) (any, error)
}
