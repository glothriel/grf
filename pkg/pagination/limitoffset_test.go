package pagination

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestLimitOffsetPagination_Apply(t *testing.T) {
	// given
	p := &LimitOffsetPagination{}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/test?limit=20&offset=10", nil)

	db, openErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// when
	newDb := p.Apply(ctx, db)
	limitClause := newDb.Statement.Clauses["LIMIT"].Expression.(clause.Limit)

	// then
	assert.NoError(t, openErr)
	assert.Equal(t, 10, limitClause.Offset)
	assert.Equal(t, 20, *limitClause.Limit)
}

func TestLimitOffsetPaginationApplyInvalidLimit(t *testing.T) {
	// given
	p := &LimitOffsetPagination{}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/test?limit=invalid&offset=10", nil)

	db, openErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// when
	newDb := p.Apply(ctx, db)
	limitClause := newDb.Statement.Clauses["LIMIT"].Expression.(clause.Limit)

	// then
	assert.NoError(t, openErr)
	assert.Equal(t, 10, limitClause.Offset)
	assert.Nil(t, limitClause.Limit)
}

func TestLimitOffsetPaginationApplyInvalidOffset(t *testing.T) {
	// given
	p := &LimitOffsetPagination{}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/test?limit=20&offset=invalid", nil)

	db, openErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// when
	newDb := p.Apply(ctx, db)
	limitClause := newDb.Statement.Clauses["LIMIT"].Expression.(clause.Limit)

	// then
	assert.NoError(t, openErr)
	assert.Equal(t, 0, limitClause.Offset)
	assert.Equal(t, 20, *limitClause.Limit)
}

func TestLimitOffsetPaginationFormat(t *testing.T) {
	// given
	p := &LimitOffsetPagination{}
	entities := []any{"test"}

	// when
	formattedEntities, err := p.Format(&gin.Context{}, entities)

	// then
	assert.NoError(t, err)
	assert.Equal(t, entities, formattedEntities)
}
