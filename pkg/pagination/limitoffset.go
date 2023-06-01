package pagination

import (
	"strconv"

	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LimitOffsetPagination struct {
}

func (p *LimitOffsetPagination) Apply(c *grfctx.Context, db *gorm.DB) *gorm.DB {
	if c.Gin.Query("limit") != "" {
		limit, conversionErr := strconv.Atoi(c.Gin.Query("limit"))
		if conversionErr != nil {
			logrus.Debug("Failed to convert limit to int in LimitOffsetPagination")
			return db
		}
		db = db.Limit(limit)
	}
	if c.Gin.Query("offset") != "" {
		offset, conversionErr := strconv.Atoi(c.Gin.Query("offset"))
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
