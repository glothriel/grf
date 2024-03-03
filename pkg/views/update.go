package views

import (
	"net/http"
	"reflect"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/sirupsen/logrus"
)

func UpdateModelViewSetFunc[Model any](idf IDFunc, qd queries.Driver[Model], serializer serializers.Serializer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var parsedBody map[string]any
		if parseErr := ctx.ShouldBindJSON(&parsedBody); parseErr != nil {
			WriteError(ctx, parseErr)
			return
		}

		updates, idEnrichErr := enrichBodyWithID[Model](ctx, hasNumericID[Model](), idf, parsedBody)
		if idEnrichErr != nil {
			WriteError(ctx, idEnrichErr)
			return
		}

		effectiveSerializer := serializer
		incomingIntVal, fromRawErr := effectiveSerializer.ToInternalValue(updates, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		oldIntVal, oldErr := qd.CRUD().Retrieve(ctx, idf(ctx))
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
		updatedIntVal, updateErr := qd.CRUD().Update(
			ctx, oldIntVal, newIntVal, idf(ctx),
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

func enrichBodyWithID[Model any](ctx *gin.Context, isNumeric bool, idf IDFunc, b map[string]any) (map[string]any, error) {
	idFromURLStr := idf(ctx)
	if !isNumeric {
		if idFromBody, ok := b["id"]; ok {
			if idFromBody != idFromURLStr {
				return nil, &serializers.ValidationError{
					FieldErrors: map[string][]string{
						"id": {"id in body does not match id in url"},
					},
				}
			}
		} else {
			b["id"] = idFromURLStr
		}
		return b, nil
	}

	idFromUrlFloat, convertErr := strconv.ParseFloat(idFromURLStr, 64)
	if convertErr != nil {
		return nil, convertErr
	}
	if idFromBody, ok := b["id"]; ok {
		if idFromBody != idFromUrlFloat {
			return nil, &serializers.ValidationError{
				FieldErrors: map[string][]string{
					"id": {"id in body does not match id in url"},
				},
			}
		}
	} else {
		b["id"] = idFromUrlFloat
	}
	return b, nil
}

func hasNumericID[Model any]() bool {
	var m Model
	intVal := models.AsInternalValue(m)
	_, ok := intVal["id"]
	if !ok {
		logrus.Panicf("Missing id field on model %T", m)
	}
	return slices.Contains([]string{"int", "int64", "uint", "uint64"}, reflect.TypeOf(intVal["id"]).Kind().String())
}
