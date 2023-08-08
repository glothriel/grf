package views

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

func TestWriteError(t *testing.T) {
	type errorTest struct {
		name     string
		err      error
		expected int
	}

	tests := []errorTest{
		{
			name:     "validation error",
			err:      &serializers.ValidationError{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "not found error",
			err:      common.ErrorNotFound,
			expected: http.StatusNotFound,
		},
		{
			name:     "syntax error",
			err:      &json.SyntaxError{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "unexpected EOF error",
			err:      io.ErrUnexpectedEOF,
			expected: http.StatusBadRequest,
		},
		{
			name:     "EOF error",
			err:      io.EOF,
			expected: http.StatusBadRequest,
		},
		{
			name:     "generic error",
			err:      errors.New("Some generic unknown error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, r := gin.CreateTestContext(httptest.NewRecorder())
			r.GET("/test", func(c *gin.Context) {
				WriteError(c, tt.err)
			})

			request, requestErr := http.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			assert.NoError(t, requestErr)
			assert.Equal(t, tt.expected, w.Code)
		})
	}
}
