package views

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"
	"github.com/glothriel/grf/pkg/queries"
)

type ViewRoute struct {
	Method       string
	RelativePath string
	Handler      func(*gin.Context)
}

type View struct {
	path          string
	getHandler    func(*gin.Context)
	postHandler   func(*gin.Context)
	putHandler    func(*gin.Context)
	deleteHandler func(*gin.Context)
	patchHandler  func(*gin.Context)
	extraRoutes   []*ViewRoute
	authenticator authentication.Authentication

	middleware []gin.HandlerFunc
}

func (v *View) Get(h func(*gin.Context)) *View {
	v.getHandler = h
	return v
}

func (v *View) Post(h func(*gin.Context)) *View {
	v.postHandler = h
	return v
}

func (v *View) Put(h func(*gin.Context)) *View {
	v.putHandler = h
	return v
}

func (v *View) Delete(h func(*gin.Context)) *View {
	v.deleteHandler = h
	return v
}

func (v *View) Patch(h func(*gin.Context)) *View {
	v.patchHandler = h
	return v
}

func (v *View) WithRoute(
	route *ViewRoute,
) *View {
	v.extraRoutes = append(v.extraRoutes, route)
	return v
}

func (v *View) AddMiddleware(m ...gin.HandlerFunc) *View {
	v.middleware = append(v.middleware, m...)
	return v
}

func (v *View) Register(r *gin.Engine) {
	rg := r.Group(v.path, v.middleware...)
	if v.getHandler != nil {
		rg.GET("", v.getHandler)
	}
	if v.postHandler != nil {
		rg.POST("", v.postHandler)
	}
	if v.putHandler != nil {
		rg.PUT("", v.putHandler)
	}
	if v.deleteHandler != nil {
		rg.DELETE("", v.deleteHandler)
	}
	if v.patchHandler != nil {
		rg.PATCH("", v.patchHandler)
	}
	for _, extraAction := range v.extraRoutes {
		rg.Handle(extraAction.Method, extraAction.RelativePath, extraAction.Handler)
	}
}

func NewView[Model any](path string, queryDriver queries.Driver[Model]) *View {

	return &View{
		path:          path,
		getHandler:    nil,
		postHandler:   nil,
		putHandler:    nil,
		deleteHandler: nil,
		patchHandler:  nil,
		authenticator: &authentication.AnonymousUserAuthentication{},
		extraRoutes:   []*ViewRoute{},

		middleware: queryDriver.Middleware(),
	}
}
