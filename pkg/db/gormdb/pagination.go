package gormdb

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Pagination interface {
	Apply(*gin.Context, *gorm.DB) *gorm.DB
	Format(*gin.Context, []any) (any, error)
}

type NoPagination struct{}

func (p *NoPagination) Apply(_ *gin.Context, db *gorm.DB) *gorm.DB {
	return db
}

func (p *NoPagination) Format(_ *gin.Context, entities []any) (any, error) {
	return entities, nil
}

type LimitOffsetPagination struct {
}

func (p *LimitOffsetPagination) Apply(c *gin.Context, db *gorm.DB) *gorm.DB {
	if c.Query("limit") != "" {
		limit, conversionErr := strconv.Atoi(c.Query("limit"))
		if conversionErr == nil {
			db = db.Limit(limit)
		} else {
			logrus.Debug("Failed to convert limit to int in LimitOffsetPagination")
		}
	}
	if c.Query("offset") != "" {
		offset, conversionErr := strconv.Atoi(c.Query("offset"))
		if conversionErr == nil {
			db = db.Offset(offset)
		} else {
			logrus.Debug("Failed to convert offset to int in LimitOffsetPagination")
		}
	}

	return db
}

func (p *LimitOffsetPagination) Format(c *gin.Context, entities []any) (any, error) {
	return entities, nil
}
