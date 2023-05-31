package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/authentication"
	"github.com/glothriel/grf/pkg/grfctx"
)

type View struct {
	path          string
	getHandler    func(*gin.Context)
	postHandler   func(*gin.Context)
	putHandler    func(*gin.Context)
	deleteHandler func(*gin.Context)
	patchHandler  func(*gin.Context)
	authenticator authentication.Authentication
	context       *grfctx.Context
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

func (v *View) Authentication(a authentication.Authentication) *View {
	v.authenticator = a
	return v
}

func (v *View) Register(r *gin.Engine) {
	r.GET(v.path, v.authenticated(v.getHandler))
	r.POST(v.path, v.authenticated(v.postHandler))
	r.PUT(v.path, v.authenticated(v.putHandler))
	r.DELETE(v.path, v.authenticated(v.deleteHandler))
	r.PATCH(v.path, v.authenticated(v.patchHandler))
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

func NewView(path string) *View {
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
	}
}
