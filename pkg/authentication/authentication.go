package authentication

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Authentication interface {
	Authenticate(*gin.Context) (bool, error)
}

type AnonymousUserAuthentication struct{}

func (a *AnonymousUserAuthentication) Authenticate(c *gin.Context) (bool, error) {
	c.Set("user", &User{
		Name:  "Anonymous",
		Email: "anonymous@localhost",
	})
	return true, nil
}

func CurrentUser(c *gin.Context) (*User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, errors.New("No user was authenticated, please use the correct authentication middleware")
	}
	return user.(*User), nil
}
