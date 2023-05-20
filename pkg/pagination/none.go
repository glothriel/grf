package pagination

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NoPagination struct{}

func (p *NoPagination) Apply(_ *gin.Context, db *gorm.DB) *gorm.DB {
	return db
}

func (p *NoPagination) Format(entities []any) (any, error) {
	return entities, nil
}
