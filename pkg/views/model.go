package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"
	"github.com/glothriel/grf/pkg/queries/crud"

	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/types"
)

type ModelViewSettings[Model any] struct {
	DefaultSerializer  serializers.Serializer
	ListSerializer     serializers.Serializer
	RetrieveSerializer serializers.Serializer
	UpdateSerializer   serializers.Serializer
	CreateSerializer   serializers.Serializer
	DestroySerializer  serializers.Serializer

	IDFunc      func(*gin.Context) string
	QueryDriver queries.Driver[Model]
}

func DefaultModelViewSettings[Model any](queryDriver queries.Driver[Model]) ModelViewSettings[Model] {
	return ModelViewSettings[Model]{
		DefaultSerializer: &serializers.MissingSerializer[Model]{},
		QueryDriver:       queryDriver,
		IDFunc:            IDFromQueryParamIDFunc[Model],
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
	DestroyFunc  HandlerFactoryFunc[Model]
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
	if v.DestroyFunc != nil {
		v.View.Delete(v.DestroyFunc(v.Settings))
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

func (v *ModelView[Model]) WithDestroySerializer(serializer serializers.Serializer) *ModelView[Model] {
	v.Settings.DestroySerializer = serializer
	return v
}

func (v *ModelView[Model]) WithAuthentication(authenticator authentication.Authentication) *ModelView[Model] {
	v.View.Authentication(authenticator)
	return v
}

func (v *ModelView[Model]) WithFieldTypeMapper(fieldTypeMapper *types.FieldTypeMapper) *ModelView[Model] {
	return v
}

func (v *ModelView[Model]) OnCreate(modFunc func(c crud.CreateQueryFunc) crud.CreateQueryFunc) *ModelView[Model] {
	v.Settings.QueryDriver.CRUD().WithCreate(modFunc(v.Settings.QueryDriver.CRUD().Create))
	return v
}

func (v *ModelView[Model]) OnUpdate(modFunc func(u crud.UpdateQueryFunc) crud.UpdateQueryFunc) *ModelView[Model] {
	v.Settings.QueryDriver.CRUD().WithUpdate(modFunc(v.Settings.QueryDriver.CRUD().Update))
	return v
}

func (v *ModelView[Model]) OnDestroy(modFunc func(d crud.DestroyQueryFunc) crud.DestroyQueryFunc) *ModelView[Model] {
	v.Settings.QueryDriver.CRUD().WithDestroy(modFunc(v.Settings.QueryDriver.CRUD().Destroy))
	return v
}

func NewListCreateModelView[Model any](path string, queryDriver queries.Driver[Model]) *ModelView[Model] {
	return &ModelView[Model]{
		View:       NewView(path, queryDriver),
		Settings:   DefaultModelViewSettings(queryDriver),
		ListFunc:   ListModelFunc[Model],
		CreateFunc: CreateModelFunc[Model],
	}
}

func NewRetrieveUpdateDestroyModelView[Model any](path string, queryDriver queries.Driver[Model]) *ModelView[Model] {
	return &ModelView[Model]{
		View:         NewView(path, queryDriver),
		Settings:     DefaultModelViewSettings(queryDriver),
		RetrieveFunc: RetrieveModelFunc[Model],
		UpdateFunc:   UpdateModelFunc[Model],
		DestroyFunc:  DestroyModelFunc[Model],
	}
}

func IDFromQueryParamIDFunc[Model any](ctx *gin.Context) string {
	return ctx.Param("id")
}
