package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
)

func UpdateModelFunc[Model any](modelSettings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updates map[string]any
		if parseErr := ctx.ShouldBindJSON(&updates); parseErr != nil {
			WriteError(ctx, parseErr)
		}
		if idFromBody, ok := updates["id"]; ok {
			if idFromBody != modelSettings.IDFunc(ctx) {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": "id in body does not match id in url",
				})
				return
			}
		} else {
			updates["id"] = modelSettings.IDFunc(ctx)
		}
		effectiveSerializer := modelSettings.UpdateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelSettings.DefaultSerializer
		}
		incomingIntVal, fromRawErr := effectiveSerializer.ToInternalValue(updates, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		oldIntVal, oldErr := modelSettings.QueryDriver.CRUD().Retrieve(ctx, modelSettings.IDFunc(ctx))
		if oldErr != nil {
			WriteError(ctx, oldErr)
			return
		}
		newIntVal := models.InternalValue{}
		for k, v := range oldIntVal {
			newIntVal[k] = v
		}
		for k, v := range incomingIntVal {
			newIntVal[k] = v
		}
		updatedIntVal, updateErr := modelSettings.QueryDriver.CRUD().Update(
			ctx, oldIntVal, newIntVal, modelSettings.IDFunc(ctx),
		)
		if updateErr != nil {
			WriteError(ctx, updateErr)
			return
		}
		rawElement, toRawErr := effectiveSerializer.ToRepresentation(updatedIntVal, ctx)
		if toRawErr != nil {
			WriteError(ctx, toRawErr)
			return
		}
		ctx.JSON(http.StatusOK, rawElement)
	}
}
