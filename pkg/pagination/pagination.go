package pagination

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"gorm.io/gorm"
)

type Pagination interface {
	Apply(*grfctx.Context, *gorm.DB) *gorm.DB
	Format([]any) (any, error)
}
