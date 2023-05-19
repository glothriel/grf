package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/gin-rest-framework/pkg/authentication"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/pagination"
	"github.com/glothriel/gin-rest-framework/pkg/serializers"
	"github.com/glothriel/gin-rest-framework/pkg/types"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QueryModFunc func(*gin.Context, *gorm.DB) *gorm.DB

func QueryModPassThrough(ctx *gin.Context, db *gorm.DB) *gorm.DB {
	return db
}

func QueryModOrderBy(order string) QueryModFunc {
	return func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
		return db.Order(order)
	}
}

type ModelViewContext[Model any] struct {
	Serializer serializers.Serializer[Model]

	ListSerializer     serializers.Serializer[Model]
	RetrieveSerializer serializers.Serializer[Model]
	UpdateSerializer   serializers.Serializer[Model]
	CreateSerializer   serializers.Serializer[Model]
	DeleteSerializer   serializers.Serializer[Model]

	Pagination      pagination.Pagination
	Filter          QueryModFunc
	OrderBy         QueryModFunc
	IDFunc          func(ModelViewContext[Model], *gin.Context) string
	DB              *gorm.DB
	FieldTypeMapper *types.FieldTypeMapper
	FieldTypes      map[string]string
}

func (mvc ModelViewContext[Model]) DBSession() *gorm.DB {
	return mvc.DB.Session(&gorm.Session{})
}

func NewDefaultModelViewContext[Model any](DB *gorm.DB) ModelViewContext[Model] {
	var m Model
	return ModelViewContext[Model]{
		Serializer:      &serializers.MissingSerializer[Model]{},
		Pagination:      &pagination.NoPagination{},
		Filter:          QueryModPassThrough,
		OrderBy:         QueryModPassThrough,
		DB:              DB,
		IDFunc:          IDFromQueryParamIDFunc[Model],
		FieldTypeMapper: types.DefaultFieldTypeMapper(),
		FieldTypes:      serializers.DetectAttributes[Model](m),
	}
}

type HandlerFactoryFunc[Model any] func(ModelViewContext[Model]) gin.HandlerFunc

type ModelView[Model any] struct {
	view    *View
	Context ModelViewContext[Model]

	ListFunc     HandlerFactoryFunc[Model]
	CreateFunc   HandlerFactoryFunc[Model]
	RetrieveFunc HandlerFactoryFunc[Model]
	UpdateFunc   HandlerFactoryFunc[Model]
	DeleteFunc   HandlerFactoryFunc[Model]
}

func (v *ModelView[Model]) Register(r *gin.Engine) {
	if v.ListFunc != nil {
		v.view.Get(v.ListFunc(v.Context))
	}
	if v.CreateFunc != nil {
		v.view.Post(v.CreateFunc(v.Context))
	}
	if v.RetrieveFunc != nil {
		v.view.Get(v.RetrieveFunc(v.Context))
	}
	if v.UpdateFunc != nil {
		v.view.Put(v.UpdateFunc(v.Context))
		v.view.Patch(v.UpdateFunc(v.Context))
	}
	if v.DeleteFunc != nil {
		v.view.Delete(v.DeleteFunc(v.Context))
	}
	v.view.Register(r)
}

func (v *ModelView[Model]) WithSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.Serializer = serializer
	return v
}

func (v *ModelView[Model]) WithListSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.ListSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithRetrieveSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.RetrieveSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithUpdateSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.UpdateSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithCreateSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.CreateSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithDeleteSerializer(serializer serializers.Serializer[Model]) *ModelView[Model] {
	v.Context.DeleteSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithPagination(pagination pagination.Pagination) *ModelView[Model] {
	v.Context.Pagination = pagination
	return v
}

func (v *ModelView[Model]) WithFilter(filter func(*gin.Context, *gorm.DB) *gorm.DB) *ModelView[Model] {
	v.Context.Filter = filter
	return v
}

func (v *ModelView[Model]) WithOrderBy(order string) *ModelView[Model] {
	v.Context.OrderBy = QueryModOrderBy(order)
	return v
}

func (v *ModelView[Model]) WithAuthentication(authenticator authentication.Authentication) *ModelView[Model] {
	v.view.Authentication(authenticator)
	return v
}

func (v *ModelView[Model]) WithFieldTypeMapper(fieldTypeMapper *types.FieldTypeMapper) *ModelView[Model] {
	v.Context.FieldTypeMapper = fieldTypeMapper
	return v
}

func (v *ModelView[Model]) WithListHandlerFactoryFunc(factory HandlerFactoryFunc[Model]) *ModelView[Model] {
	v.ListFunc = factory
	return v
}

func (v *ModelView[Model]) WithCreateHandlerFactoryFunc(factory HandlerFactoryFunc[Model]) *ModelView[Model] {
	v.CreateFunc = factory
	return v
}

func (v *ModelView[Model]) WithRetrieveHandlerFactoryFunc(factory HandlerFactoryFunc[Model]) *ModelView[Model] {
	v.RetrieveFunc = factory
	return v
}

func (v *ModelView[Model]) WithUpdateHandlerFactoryFunc(factory HandlerFactoryFunc[Model]) *ModelView[Model] {
	v.UpdateFunc = factory
	return v
}

func (v *ModelView[Model]) WithDeleteHandlerFactoryFunc(factory HandlerFactoryFunc[Model]) *ModelView[Model] {
	v.DeleteFunc = factory
	return v
}

func NewListCreateModelView[Model any](path string, db *gorm.DB) *ModelView[Model] {
	var model Model
	return &ModelView[Model]{
		view:       NewView(path),
		Context:    NewDefaultModelViewContext[Model](db.Model(&model)),
		ListFunc:   ListModelView[Model],
		CreateFunc: CreateModelView[Model],
	}
}

func NewRetrieveUpdateDeleteModelView[Model any](path string, db *gorm.DB) *ModelView[Model] {
	var model Model
	return &ModelView[Model]{
		view:         NewView(path),
		Context:      NewDefaultModelViewContext[Model](db.Model(&model)),
		RetrieveFunc: RetrieveModelView[Model],
		UpdateFunc:   UpdateModelView[Model],
		DeleteFunc:   DeleteModelView[Model],
	}
}

func ListModelView[Model any](modelCtx ModelViewContext[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var entities []map[string]interface{}
		modelCtx.Pagination.Apply(
			ctx,
			modelCtx.OrderBy(ctx, modelCtx.Filter(ctx, modelCtx.DBSession())),
		).Find(&entities)
		rawElements := []interface{}{}
		effectiveSerializer := modelCtx.ListSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelCtx.Serializer
		}
		for _, entity := range entities {
			internalValue, internalValueErr := effectiveSerializer.FromDB(entity)
			if internalValueErr != nil {
				WriteError(ctx, internalValueErr)
				return
			}
			rawElement, toRawErr := effectiveSerializer.ToRepresentation(
				internalValue,
			)

			if toRawErr != nil {
				ctx.JSON(500, gin.H{
					"message": toRawErr.Error(),
				})
				return
			}
			rawElements = append(rawElements, rawElement)
		}
		ctx.JSON(200, rawElements)
	}
}

func CreateModelView[Model any](modelCtx ModelViewContext[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rawElement map[string]interface{}
		if err := ctx.ShouldBindJSON(&rawElement); err != nil {
			ctx.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		effectiveSerializer := modelCtx.CreateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelCtx.Serializer
		}
		internalValue, fromRawErr := effectiveSerializer.ToInternalValue(rawElement)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		entity, asModelErr := internalValue.AsModel()
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		// Gorm supports creating rows using maps, but we cannot use that, because in that case
		// Gorm won't execute hooks. UUID-based PKs require a hook to be executed. That's why we
		// convert the map to a model and execute the query.
		createErr := modelCtx.DBSession().Create(&entity).Error
		if createErr != nil {
			WriteError(ctx, createErr)
			return
		}
		internalValue, internalValueErr := models.InternalValueFromModel(entity)
		if internalValueErr != nil {
			WriteError(ctx, internalValueErr)
			return
		}
		representation, serializeErr := effectiveSerializer.ToRepresentation(internalValue)
		if serializeErr != nil {
			ctx.JSON(500, gin.H{
				"message": serializeErr.Error(),
			})
			return
		}
		ctx.JSON(201, representation)
	}
}

func RetrieveModelView[Model any](modelCtx ModelViewContext[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var entity Model
		if err := modelCtx.Filter(ctx, modelCtx.DBSession().First(&entity, "id = ?", ctx.Param("id"))).Error; err != nil {
			ctx.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		effectiveSerializer := modelCtx.RetrieveSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelCtx.Serializer
		}

		internalValue, internalValueErr := models.InternalValueFromModel(entity)
		if internalValueErr != nil {
			WriteError(ctx, internalValueErr)
			return
		}
		rawElement, toRawErr := effectiveSerializer.ToRepresentation(internalValue)
		if toRawErr != nil {
			ctx.JSON(500, gin.H{
				"message": toRawErr.Error(),
			})
			return
		}
		ctx.JSON(200, rawElement)
	}
}

func UpdateModelView[Model any](modelCtx ModelViewContext[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updates map[string]interface{}
		if err := ctx.ShouldBindJSON(&updates); err != nil {
			ctx.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		effectiveSerializer := modelCtx.UpdateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = modelCtx.Serializer
		}
		intVal, fromRawErr := effectiveSerializer.ToInternalValue(updates)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		intVal.Map["id"] = ctx.Param("id")
		idIVFunc, intValErr := modelCtx.FieldTypeMapper.ToInternalValue(
			modelCtx.FieldTypes["id"],
		)
		if intValErr != nil {
			WriteError(ctx, intValErr)
			return
		}
		internalID, iDErr := idIVFunc(intVal.Map["id"])
		if iDErr != nil {
			WriteError(ctx, iDErr)
			return
		}
		intVal.Map["id"] = internalID
		entity, asModelErr := intVal.AsModel()
		if asModelErr != nil {
			WriteError(ctx, asModelErr)
			return
		}
		updateErr := modelCtx.DBSession().Model(&entity).Updates(entity).Error
		if updateErr != nil {
			WriteError(ctx, updateErr)
			return
		}
		var updatedMap map[string]interface{}
		if err := modelCtx.Filter(ctx, modelCtx.DBSession().First(&updatedMap, "id = ?", ctx.Param("id"))).Error; err != nil {
			ctx.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		rawElement, toRawErr := effectiveSerializer.ToRepresentation(&models.InternalValue[Model]{Map: updatedMap})
		if toRawErr != nil {
			ctx.JSON(500, gin.H{
				"message": toRawErr.Error(),
			})
			return
		}
		ctx.JSON(200, rawElement)
	}
}

func DeleteModelView[Model any](modelCtx ModelViewContext[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var entity Model
		deleteErr := modelCtx.DBSession().Delete(&entity, "id = ?", ctx.Param("id")).Error
		if deleteErr != nil {
			ctx.JSON(500, gin.H{
				"message": deleteErr.Error(),
			})
			return
		}
		ctx.JSON(204, nil)
	}
}

func WriteError(ctx *gin.Context, err error) {
	ve, isValidationErr := err.(*serializers.ValidationError)
	if isValidationErr {
		ctx.JSON(400, gin.H{
			"errors": ve.FieldErrors,
		})
		return
	}
	logrus.Error(err)
	ctx.JSON(500, "Internal Server Error")
}

func IDFromQueryParamIDFunc[Model any](modelCtx ModelViewContext[Model], ctx *gin.Context) string {
	return ctx.Param("id")
}
