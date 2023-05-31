package serializers

import (
	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/glothriel/grf/pkg/models"
)

type Representation map[string]any

type Serializer interface {
	ToInternalValue(map[string]any, *grfctx.Context) (models.InternalValue, error)
	FromDB(map[string]any, *grfctx.Context) (models.InternalValue, error)
	ToRepresentation(models.InternalValue, *grfctx.Context) (Representation, error)
}
