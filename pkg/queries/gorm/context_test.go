package gorm

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCtxGetGorm(t *testing.T) {
	// given
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when + then
	assert.Panics(t, func() { CtxGetGorm(ctx) })
}

func TestCtxGetGormInvalidType(t *testing.T) {
	// given
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// when
	ctx.Set("db:gorm", "huehue")

	// then
	assert.Panics(t, func() { CtxGetGorm(ctx) })
}
