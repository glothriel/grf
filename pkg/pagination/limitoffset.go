package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LimitOffsetPagination struct {
}

func (p *LimitOffsetPagination) Apply(c *gin.Context, db *gorm.DB) *gorm.DB {
	if c.Query("limit") != "" {
		limit, conversionErr := strconv.Atoi(c.Query("limit"))
		if conversionErr != nil {
			logrus.Debug("Failed to convert limit to int in LimitOffsetPagination")
			return db
		}
		db = db.Limit(limit)
	}
	if c.Query("offset") != "" {
		offset, conversionErr := strconv.Atoi(c.Query("offset"))
		if conversionErr != nil {
			logrus.Debug("Failed to convert offset to int in LimitOffsetPagination")
			return db
		}
		db = db.Offset(offset)
	}

	return db
}

func (p *LimitOffsetPagination) Format(entities []any) (any, error) {
	return entities, nil
}
