package views

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

type MockModel struct {
	Foo string `json:"foo"`
}

var createModelViewTests = []struct {
	name             string
	json             string
	wantStatus       int
	wantBodyContains string
}{
	{
		name:             "Valid JSON",
		json:             `{"foo": "bar"}`,
		wantStatus:       201,
		wantBodyContains: `{"foo":"bar"}`,
	},
	{
		name:             "Invalid JSON",
		json:             `{"foo": "bar"`,
		wantStatus:       400,
		wantBodyContains: "EOF",
	},
	{
		name:             "Invalid Fields",
		json:             `{"bar": "baz"}`,
		wantStatus:       400,
		wantBodyContains: "accepted fields",
	},
}

func TestCreateModelView(t *testing.T) {
	for _, tt := range createModelViewTests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			settings := ModelViewSettings[MockModel]{
				DefaultSerializer: serializers.NewModelSerializer[MockModel](),
				Database:          db.Dummy(MockModel{Foo: "bar"}),
			}
			_, r := gin.CreateTestContext(httptest.NewRecorder())
			r.POST("/users", CreateModelFunc(settings))
			request, requestErr := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.json))
			w := httptest.NewRecorder()

			// when
			r.ServeHTTP(w, request)

			// then
			a, readErr := io.ReadAll(w.Body)
			assert.NoError(t, requestErr)
			assert.NoError(t, readErr)
			assert.Contains(t, string(a), tt.wantBodyContains)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
