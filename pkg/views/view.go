package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"
	"github.com/glothriel/grf/pkg/db"
)

type View struct {
	path          string
	getHandler    func(*gin.Context)
	postHandler   func(*gin.Context)
	putHandler    func(*gin.Context)
	deleteHandler func(*gin.Context)
	patchHandler  func(*gin.Context)
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

func (v *View) AddMiddleware(m ...gin.HandlerFunc) *View {
	v.middleware = append(v.middleware, m...)
	return v
}

func (v *View) Authentication(a authentication.Authentication) *View {
	v.authenticator = a
	return v
}

func (v *View) Register(r *gin.Engine) {
	rg := r.Group(v.path, v.middleware...)
	rg.GET("", v.authenticated(v.getHandler))
	rg.POST("", v.authenticated(v.postHandler))
	rg.PUT("", v.authenticated(v.putHandler))
	rg.DELETE("", v.authenticated(v.deleteHandler))
	rg.PATCH("", v.authenticated(v.patchHandler))
}

func (v *View) authenticated(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := v.authenticator.Authenticate(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		if !allowed {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			return
		}
		h(c)
	}
}

func NewView[Model any](path string, database db.Database[Model]) *View {
	defaultHandler := func(ctx *gin.Context) {
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{
			"message": "Not allowed",
		})
	}

	return &View{
		path:          path,
		getHandler:    defaultHandler,
		postHandler:   defaultHandler,
		putHandler:    defaultHandler,
		deleteHandler: defaultHandler,
		patchHandler:  defaultHandler,
		authenticator: &authentication.AnonymousUserAuthentication{},

		middleware: database.Middleware(),
	}
}
