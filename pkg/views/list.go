package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/grfctx"
)

func ListModelFunc[Model any](modelSettings ModelViewSettings[Model]) HandlerFunc {
	return func(ctx *grfctx.Context) {

		var entities []map[string]any
		modelSettings.Pagination.Apply(
			ctx,
			modelSettings.OrderBy(ctx, modelSettings.Filter(ctx, db.CtxNewQuery[Model](ctx))),
		).Find(&entities)
		representationItems := []any{}
		effectiveSerializer := modelSettings.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		for _, entity := range entities {
			internalValue, internalValueErr := effectiveSerializer.FromDB(entity, ctx)
			if internalValueErr != nil {
				WriteError(ctx.Gin, internalValueErr)
				return
			}
			rawElement, toRawErr := effectiveSerializer.ToRepresentation(
				internalValue, ctx,
			)

			if toRawErr != nil {
				ctx.Gin.JSON(500, gin.H{
					"message": toRawErr.Error(),
				})
				return
			}
			representationItems = append(representationItems, rawElement)
		}
		ctx.Gin.JSON(200, representationItems)
	}
}
