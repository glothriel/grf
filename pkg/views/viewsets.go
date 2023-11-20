package views

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/queries/crud"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/types"
)

const (
	ActionCreate = iota
	ActionUpdate
	ActionDestroy
	ActionList
	ActionRetrieve
)

type ActionID int

type ViewSet[Model any] struct {
	Path        string
	IDFunc      IDFunc
	QueryDriver queries.Driver[Model]

	ListAction     *ViewSetAction[Model]
	CreateAction   *ViewSetAction[Model]
	RetrieveAction *ViewSetAction[Model]
	UpdateAction   *ViewSetAction[Model]
	DestroyAction  *ViewSetAction[Model]

	DefaultSerializer serializers.Serializer

	ListCreateView            *View
	RetrieveUpdateDestroyView *View
}

func (v *ViewSet[Model]) Register(r *gin.Engine) {
	if v.ListAction != nil {
		v.ListCreateView.Get(v.ListAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.ListAction.Serializer))
	}
	if v.CreateAction != nil {
		v.ListCreateView.Post(v.CreateAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.CreateAction.Serializer))
	}
	if v.RetrieveAction != nil {
		v.RetrieveUpdateDestroyView.Get(v.RetrieveAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.RetrieveAction.Serializer))
	}
	if v.UpdateAction != nil {
		v.RetrieveUpdateDestroyView.Put(v.UpdateAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.UpdateAction.Serializer))
		v.RetrieveUpdateDestroyView.Patch(v.UpdateAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.UpdateAction.Serializer))
	}
	if v.DestroyAction != nil {
		v.RetrieveUpdateDestroyView.Delete(v.DestroyAction.ViewsetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.DestroyAction.Serializer))
	}
	v.ListCreateView.Register(r)
	v.RetrieveUpdateDestroyView.Register(r)
}

func (v *ViewSet[Model]) WithSerializer(serializer serializers.Serializer) *ViewSet[Model] {
	v.DefaultSerializer = serializer
	return v.WithListSerializer(serializer).WithRetrieveSerializer(serializer).WithUpdateSerializer(serializer).WithCreateSerializer(serializer).WithDestroySerializer(serializer)
}

func (v *ViewSet[Model]) WithListSerializer(serializer serializers.Serializer) *ViewSet[Model] {
	if v.ListAction != nil {
		v.ListAction.Serializer = serializer
	}
	return v
}

func (v *ViewSet[Model]) WithRetrieveSerializer(serializer serializers.Serializer) *ViewSet[Model] {
	if v.RetrieveAction != nil {
		v.RetrieveAction.Serializer = serializer
	}
	return v
}

func (v *ViewSet[Model]) WithUpdateSerializer(serializer serializers.Serializer) *ViewSet[Model] {
	if v.UpdateAction != nil {
		v.UpdateAction.Serializer = serializer
	}
	return v
}

func (v *ViewSet[Model]) WithCreateSerializer(serializer serializers.Serializer) *ViewSet[Model] {
	if v.CreateAction != nil {
		v.CreateAction.Serializer = serializer
	}
	return v
}

func (v *ViewSet[Model]) WithDestroySerializer(serializer serializers.Serializer) *ViewSet[Model] {
	if v.DestroyAction != nil {
		v.DestroyAction.Serializer = serializer
	}
	return v
}

func (v *ViewSet[Model]) WithFieldTypeMapper(fieldTypeMapper *types.FieldTypeMapper) *ViewSet[Model] {
	return v
}

func (v *ViewSet[Model]) WithList(handlerFactoryFunc ViewsetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.ListAction == nil {
		v.ListAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.ListCreateView,
			ViewsetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.ListAction.ViewsetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithCreate(handlerFactoryFunc ViewsetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.CreateAction == nil {
		v.CreateAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.ListCreateView,
			ViewsetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.CreateAction.ViewsetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithRetrieve(handlerFactoryFunc ViewsetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.RetrieveAction == nil {
		v.RetrieveAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewsetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.RetrieveAction.ViewsetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithUpdate(handlerFactoryFunc ViewsetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.UpdateAction == nil {
		v.UpdateAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewsetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.UpdateAction.ViewsetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithDestroy(handlerFactoryFunc ViewsetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.DestroyAction == nil {
		v.DestroyAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewsetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.DestroyAction.ViewsetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithActions(actions ...ActionID) *ViewSet[Model] {
	for _, action := range actions {
		switch action {
		case ActionCreate:
			v.WithCreate(CreateModelViewSetFunc[Model])
		case ActionUpdate:
			v.WithUpdate(UpdateModelViewSetFunc[Model])
		case ActionDestroy:
			v.WithDestroy(DestroyModelViewSetFunc[Model])
		case ActionList:
			v.WithList(ListModelViewSetFunc[Model])
		case ActionRetrieve:
			v.WithRetrieve(RetrieveModelViewSetFunc[Model])
		}
	}
	return v
}

func (v *ViewSet[Model]) OnCreate(modFunc func(c crud.CreateQueryFunc) crud.CreateQueryFunc) *ViewSet[Model] {
	v.QueryDriver.CRUD().WithCreate(modFunc(v.QueryDriver.CRUD().Create))
	return v
}

func (v *ViewSet[Model]) OnUpdate(modFunc func(u crud.UpdateQueryFunc) crud.UpdateQueryFunc) *ViewSet[Model] {
	v.QueryDriver.CRUD().WithUpdate(modFunc(v.QueryDriver.CRUD().Update))
	return v
}

func (v *ViewSet[Model]) OnDestroy(modFunc func(d crud.DestroyQueryFunc) crud.DestroyQueryFunc) *ViewSet[Model] {
	v.QueryDriver.CRUD().WithDestroy(modFunc(v.QueryDriver.CRUD().Destroy))
	return v
}

func NewModelViewSet[Model any](path string, queryDriver queries.Driver[Model]) *ViewSet[Model] {
	return NewViewSet(path, queryDriver).WithActions(ActionCreate, ActionUpdate, ActionDestroy, ActionList, ActionRetrieve)
}

func NewViewSet[Model any](path string, queryDriver queries.Driver[Model]) *ViewSet[Model] {
	return &ViewSet[Model]{
		Path:                      path,
		QueryDriver:               queryDriver,
		IDFunc:                    IDFromQueryParamIDFunc,
		DefaultSerializer:         serializers.NewModelSerializer[Model](),
		ListCreateView:            NewView(path, queryDriver),
		RetrieveUpdateDestroyView: NewView(fmt.Sprintf("%s/:id", path), queryDriver),
	}
}

type ViewSetAction[Model any] struct {
	Path string

	View                      *View
	ViewsetHandlerFactoryFunc ViewsetHandlerFactoryFunc[Model]
	Serializer                serializers.Serializer
	QueryDriver               queries.Driver[Model]
}
