package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"

	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/types"
)

type ModelViewSettings[Model any] struct {
	DefaultSerializer  serializers.Serializer
	ListSerializer     serializers.Serializer
	RetrieveSerializer serializers.Serializer
	UpdateSerializer   serializers.Serializer
	CreateSerializer   serializers.Serializer
	DeleteSerializer   serializers.Serializer

	IDFunc   func(*gin.Context) any
	Database db.Database[Model]
}

func DefaultModelViewSettings[Model any](database db.Database[Model]) ModelViewSettings[Model] {
	return ModelViewSettings[Model]{
		DefaultSerializer: &serializers.MissingSerializer[Model]{},
		Database:          database,
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

func (v *ModelView[Model]) WithRetrieveQuery(f func(previous db.RetrieveQueryFunc[Model]) db.RetrieveQueryFunc[Model]) *ModelView[Model] {
	v.Settings.Database.Queries().WithRetrieve(
		f(v.Settings.Database.Queries().Retrieve),
	)
	return v
}

func (v *ModelView[Model]) WithListQuery(f func(previous db.ListQueryFunc[Model]) db.ListQueryFunc[Model]) *ModelView[Model] {
	v.Settings.Database.Queries().WithList(
		f(v.Settings.Database.Queries().List),
	)
	return v
}

func (v *ModelView[Model]) WithCreateQuery(f func(previous db.CreateQueryFunc[Model]) db.CreateQueryFunc[Model]) *ModelView[Model] {
	v.Settings.Database.Queries().WithCreate(
		f(v.Settings.Database.Queries().Create),
	)
	return v
}

func (v *ModelView[Model]) WithUpdateQuery(f func(previous db.UpdateQueryFunc[Model]) db.UpdateQueryFunc[Model]) *ModelView[Model] {
	v.Settings.Database.Queries().WithUpdate(
		f(v.Settings.Database.Queries().Update),
	)
	return v
}

func (v *ModelView[Model]) WithDeleteQuery(f func(previous db.DeleteQueryFunc[Model]) db.DeleteQueryFunc[Model]) *ModelView[Model] {
	v.Settings.Database.Queries().WithDelete(
		f(v.Settings.Database.Queries().Delete),
	)
	return v
}

func (v *ModelView[Model]) WithAuthentication(authenticator authentication.Authentication) *ModelView[Model] {
	v.View.Authentication(authenticator)
	return v
}

func (v *ModelView[Model]) WithFieldTypeMapper(fieldTypeMapper *types.FieldTypeMapper) *ModelView[Model] {
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

func NewListCreateModelView[Model any](path string, database db.Database[Model]) *ModelView[Model] {
	return &ModelView[Model]{
		View:       NewView(path, database),
		Settings:   DefaultModelViewSettings[Model](database),
		ListFunc:   ListModelFunc[Model],
		CreateFunc: CreateModelFunc[Model],
	}
}

func NewRetrieveUpdateDeleteModelView[Model any](path string, database db.Database[Model]) *ModelView[Model] {
	return &ModelView[Model]{
		View:         NewView(path, database),
		Settings:     DefaultModelViewSettings[Model](database),
		RetrieveFunc: RetrieveModelFunc[Model],
		UpdateFunc:   UpdateModelFunc[Model],
		DeleteFunc:   DeleteModelFunc[Model],
	}
}

func IDFromQueryParamIDFunc[Model any](ctx *gin.Context) any {
	return ctx.Param("id")
}
