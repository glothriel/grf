package serializers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingSerializer(t *testing.T) {
	// given
	ms := MissingSerializer[struct{}]{}

	// when
	_, internalValueErr := ms.ToInternalValue(map[string]any{}, nil)
	_, fromDBErr := ms.FromDB(map[string]any{}, nil)
	_, toRepresentationErr := ms.ToRepresentation(nil, nil)

	// then
	assert.Error(t, internalValueErr)
	assert.Error(t, fromDBErr)
	assert.Error(t, toRepresentationErr)

}
