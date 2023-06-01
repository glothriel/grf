package pagination

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"gorm.io/gorm"
)

type NoPagination struct{}

func (p *NoPagination) Apply(_ *grfctx.Context, db *gorm.DB) *gorm.DB {
	return db
}

func (p *NoPagination) Format(entities []any) (any, error) {
	return entities, nil
}
