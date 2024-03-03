package views

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/queries/crud"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/types"
	"github.com/sirupsen/logrus"
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

func (v *ViewSet[Model]) WithExtraAction(
	action *ExtraAction[Model],
	serializer serializers.Serializer,
	isDetail bool,
) *ViewSet[Model] {
	view := v.ListCreateView
	if isDetail {
		logrus.Error("huehueh")
		view = v.RetrieveUpdateDestroyView
	}

	view.WithRoute(&ViewRoute{
		Method:       action.Method,
		RelativePath: action.RelativePath,
		Handler:      action.Handler(v.IDFunc, v.QueryDriver, serializer),
	})
	return v
}

func (v *ViewSet[Model]) Register(r *gin.Engine) {
	if v.ListAction != nil {
		v.ListCreateView.Get(v.ListAction.ViewSetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.ListAction.Serializer))
	}
	if v.CreateAction != nil {
		v.ListCreateView.Post(v.CreateAction.ViewSetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.CreateAction.Serializer))
	}
	if v.RetrieveAction != nil {
		v.RetrieveUpdateDestroyView.Get(v.RetrieveAction.ViewSetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.RetrieveAction.Serializer))
	}
	if v.UpdateAction != nil {
		v.RetrieveUpdateDestroyView.Put(v.UpdateAction.ViewSetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.UpdateAction.Serializer))
	}
	if v.DestroyAction != nil {
		v.RetrieveUpdateDestroyView.Delete(v.DestroyAction.ViewSetHandlerFactoryFunc(v.IDFunc, v.QueryDriver, v.DestroyAction.Serializer))
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

func (v *ViewSet[Model]) WithList(handlerFactoryFunc ViewSetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.ListAction == nil {
		v.ListAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.ListCreateView,
			ViewSetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.ListAction.ViewSetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithCreate(handlerFactoryFunc ViewSetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.CreateAction == nil {
		v.CreateAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.ListCreateView,
			ViewSetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.CreateAction.ViewSetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithRetrieve(handlerFactoryFunc ViewSetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.RetrieveAction == nil {
		v.RetrieveAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewSetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.RetrieveAction.ViewSetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithUpdate(handlerFactoryFunc ViewSetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.UpdateAction == nil {
		v.UpdateAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewSetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.UpdateAction.ViewSetHandlerFactoryFunc = handlerFactoryFunc
	}
	return v
}

func (v *ViewSet[Model]) WithDestroy(handlerFactoryFunc ViewSetHandlerFactoryFunc[Model]) *ViewSet[Model] {
	if v.DestroyAction == nil {
		v.DestroyAction = &ViewSetAction[Model]{
			Path:                      v.Path,
			View:                      v.RetrieveUpdateDestroyView,
			ViewSetHandlerFactoryFunc: handlerFactoryFunc,
			Serializer:                v.DefaultSerializer,
			QueryDriver:               v.QueryDriver,
		}
	} else {
		v.DestroyAction.ViewSetHandlerFactoryFunc = handlerFactoryFunc
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

func NewViewSet[Model any](routerPath string, queryDriver queries.Driver[Model]) *ViewSet[Model] {
	// I really don't like that, the ID param is not just "id", but otherwise Gin throws a panic when
	// trying to use ViewSets on paths with multiple IDs.
	// For example /products/:product_id/photos/:id will panic if you already have /products/:id registered (product_id != id).
	// If you use /products/:id/photos/:id, Gin doesn't panic, but ofc it doesn't work as expected, ignoring the second :id.
	// I'm generating the ID param name based on the router path, so you can register (eg) "photos" viewset on
	// "/products/:product_id/photos" and Retrieve action for photos would be on "/products/:product_id/photos/:photo_id".
	// Ugly, but nothing panics and there is just no other way to go around this Gin limitation.
	var m Model
	// IDParamFunc shoud be interface with Name() and Value() so that you can easily get the ID from the code.
	idParamName := strings.ToLower(fmt.Sprintf("%s_id", reflect.TypeOf(m).Name()))
	retrieveUpdateDestroyPath := path.Join(routerPath, fmt.Sprintf(":%s", idParamName))

	return &ViewSet[Model]{
		Path:                      routerPath,
		QueryDriver:               queryDriver,
		IDFunc:                    IDFromPathParam(idParamName),
		DefaultSerializer:         serializers.NewModelSerializer[Model](),
		ListCreateView:            NewView(routerPath, queryDriver),
		RetrieveUpdateDestroyView: NewView(retrieveUpdateDestroyPath, queryDriver),
	}
}

type ViewSetAction[Model any] struct {
	Path string

	View                      *View
	ViewSetHandlerFactoryFunc ViewSetHandlerFactoryFunc[Model]
	Serializer                serializers.Serializer
	QueryDriver               queries.Driver[Model]
}
