package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FooModel struct {
	ID  uint   `json:"id" gorm:"primaryKey"`
	Foo string `json:"foo"`
}

func TestAsModel(t *testing.T) {
	// given
	internalValue := InternalValue{
		"id":  uint(1),
		"foo": "bar",
	}

	// when
	model, err := AsModel[FooModel](internalValue)

	// then
	assert.NoError(t, err)
	assert.Equal(t, FooModel{
		ID:  uint(1),
		Foo: "bar",
	}, model)
}
