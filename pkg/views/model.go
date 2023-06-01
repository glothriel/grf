package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"

	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/pagination"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/types"
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

type ModelViewSettings[Model any] struct {
	DefaultSerializer  serializers.Serializer
	ListSerializer     serializers.Serializer
	RetrieveSerializer serializers.Serializer
	UpdateSerializer   serializers.Serializer
	CreateSerializer   serializers.Serializer
	DeleteSerializer   serializers.Serializer

	Pagination      pagination.Pagination
	Filter          QueryModFunc
	OrderBy         QueryModFunc
	IDFunc          func(*gin.Context) any
	DBResolver      db.Resolver
	FieldTypeMapper *types.FieldTypeMapper
	FieldTypes      map[string]string
}

func NewDefaultModelViewContext[Model any](dbResolver db.Resolver) ModelViewSettings[Model] {
	var m Model
	return ModelViewSettings[Model]{
		DefaultSerializer: &serializers.MissingSerializer[Model]{},
		Pagination:        &pagination.NoPagination{},
		Filter:            QueryModPassThrough,
		OrderBy:           QueryModPassThrough,
		DBResolver:        dbResolver,
		IDFunc:            IDFromQueryParamIDFunc[Model],
		FieldTypeMapper:   types.DefaultFieldTypeMapper(),
		FieldTypes:        serializers.DetectAttributes(m),
	}
}

type HandlerFactoryFunc[Model any] func(ModelViewSettings[Model]) gin.HandlerFunc

type ModelView[Model any] struct {
	View     *View
	Settings ModelViewSettings[Model]

	ListFunc     HandlerFactoryFunc[Model]
	CreateFunc   HandlerFactoryFunc[Model]
	RetrieveFunc HandlerFactoryFunc[Model]
	UpdateFunc   HandlerFactoryFunc[Model]
	DeleteFunc   HandlerFactoryFunc[Model]
}

func (v *ModelView[Model]) Register(r *gin.Engine) {
	if v.ListFunc != nil {
		v.View.Get(v.ListFunc(v.Settings))
	}
	if v.CreateFunc != nil {
		v.View.Post(v.CreateFunc(v.Settings))
	}
	if v.RetrieveFunc != nil {
		v.View.Get(v.RetrieveFunc(v.Settings))
	}
	if v.UpdateFunc != nil {
		v.View.Put(v.UpdateFunc(v.Settings))
		v.View.Patch(v.UpdateFunc(v.Settings))
	}
	if v.DeleteFunc != nil {
		v.View.Delete(v.DeleteFunc(v.Settings))
	}
	v.View.Register(r)
}

func (v *ModelView[Model]) WithSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.DefaultSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithListSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.ListSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithRetrieveSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.RetrieveSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithUpdateSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.UpdateSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithCreateSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.CreateSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithDeleteSerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.DeleteSerializer = serializer
	return v
}

func (v *ModelView[Model]) WithPagination(pagination pagination.Pagination) *ModelView[Model] {
	v.Settings.Pagination = pagination
	return v
}

func (v *ModelView[Model]) WithFilter(filter QueryModFunc) *ModelView[Model] {
	v.Settings.Filter = filter
	return v
}

func (v *ModelView[Model]) WithOrderBy(order string) *ModelView[Model] {
	v.Settings.OrderBy = QueryModOrderBy(order)
	return v
}

func (v *ModelView[Model]) WithAuthentication(authenticator authentication.Authentication) *ModelView[Model] {
	v.View.Authentication(authenticator)
	return v
}

func (v *ModelView[Model]) WithFieldTypeMapper(fieldTypeMapper *types.FieldTypeMapper) *ModelView[Model] {
	v.Settings.FieldTypeMapper = fieldTypeMapper
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

func NewListCreateModelView[Model any](path string, dbResolver db.Resolver) *ModelView[Model] {
	return &ModelView[Model]{
		View:       NewView(path, dbResolver),
		Settings:   NewDefaultModelViewContext[Model](dbResolver),
		ListFunc:   ListModelFunc[Model],
		CreateFunc: CreateModelFunc[Model],
	}
}

func NewRetrieveUpdateDeleteModelView[Model any](path string, dbResolver db.Resolver) *ModelView[Model] {
	return &ModelView[Model]{
		View:         NewView(path, dbResolver),
		Settings:     NewDefaultModelViewContext[Model](dbResolver),
		RetrieveFunc: RetrieveModelFunc[Model],
		UpdateFunc:   UpdateModelFunc[Model],
		DeleteFunc:   DeleteModelFunc[Model],
	}
}

func IDFromQueryParamIDFunc[Model any](ctx *gin.Context) any {
	return ctx.Param("id")
}
