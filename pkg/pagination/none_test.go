package pagination

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNoPagination_Apply(t *testing.T) {
	// given
	p := &NoPagination{}
	initialDb := &gorm.DB{}

	// when
	db := p.Apply(&gin.Context{}, &gorm.DB{})

	// then
	assert.Equal(t, initialDb, db)
}

func TestNoPagination_Format(t *testing.T) {
	// given
	p := &NoPagination{}
	entities := []any{"test"}

	// when
	formattedEntities, err := p.Format(&gin.Context{}, entities)

	// then
	assert.NoError(t, err)
	assert.Equal(t, entities, formattedEntities)
}
