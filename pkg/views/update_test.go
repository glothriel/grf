package views

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

var updateModelViewTests = []struct {
	name                string
	json                string
	wantStatus          int
	wantBodyContainsStr string
	wantBodyJSONEquals  any
}{
	{
		name:               "Valid",
		json:               `{"foo": "baz"}`,
		wantStatus:         http.StatusOK,
		wantBodyJSONEquals: map[string]any{"id": float64(2), "foo": "baz"},
	},
	{
		name:       "Invalid JSON",
		json:       `{"foo": "bar"`,
		wantStatus: 400,
		wantBodyJSONEquals: map[string]any{
			"errors": map[string]any{
				"all": []any{"could not parse request body"},
			},
		},
	},
	{
		name:       "Invalid Fields",
		json:       `{"bar": "baz"}`,
		wantStatus: 400,
		wantBodyJSONEquals: map[string]any{
			"errors": map[string]any{
				"bar": []any{"Field `bar` is not accepted by this endpoint"},
			},
		},
	},
}

func TestUpdateModelView(t *testing.T) {
	for _, tt := range updateModelViewTests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qd := queries.InMemory(MockModel{Foo: "bar"})
			serializer := serializers.NewModelSerializer[MockModel]()

			_, r := gin.CreateTestContext(httptest.NewRecorder())

			r.POST("/foos/", CreateModelViewSetFunc(IDFromQueryParamIDFunc, qd, serializer))
			r.PUT("/foos/:id", UpdateModelViewSetFunc(IDFromQueryParamIDFunc, qd, serializer))
			createRequest, createRequestErr := http.NewRequest(http.MethodPost, "/foos/", bytes.NewBufferString(`{"foo": "bar"}`))
			updateRequest, updateRequestErr := http.NewRequest(http.MethodPut, "/foos/2", bytes.NewBufferString(tt.json))
			updateRecorder := httptest.NewRecorder()

			// when
			r.ServeHTTP(httptest.NewRecorder(), createRequest)
			r.ServeHTTP(updateRecorder, updateRequest)

			// then
			updateBody, readErr := io.ReadAll(updateRecorder.Body)
			assert.NoError(t, createRequestErr)
			assert.NoError(t, updateRequestErr)
			assert.NoError(t, readErr)
			if tt.wantBodyContainsStr != "" {
				assert.Contains(t, string(updateBody), tt.wantBodyContainsStr)
			}
			if tt.wantBodyJSONEquals != nil {
				var responseJSON any
				responseJSONErr := json.Unmarshal(updateBody, &responseJSON)
				assert.NoError(t, responseJSONErr)
				assert.Equal(t, tt.wantBodyJSONEquals, responseJSON)
			}
			assert.Equal(t, tt.wantStatus, updateRecorder.Code)
		})
	}
}
