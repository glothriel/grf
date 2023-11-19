package views

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/queries/dummy"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

type MockModel struct {
	ID  uint   `json:"id"`
	Foo string `json:"foo"`
}

var createModelViewTests = []struct {
	name                string
	json                string
	driverMod           func(*dummy.InMemoryQueryDriver[MockModel]) *dummy.InMemoryQueryDriver[MockModel]
	wantStatus          int
	wantBodyContainsStr string
	wantBodyJSONEquals  any
}{
	{
		name:               "Valid",
		json:               `{"foo": "bar"}`,
		wantStatus:         201,
		wantBodyJSONEquals: map[string]any{"id": float64(2), "foo": "bar"},
	},
	{
		name: "Driver error",
		json: `{"foo": "bar"}`,
		driverMod: func(imqd *dummy.InMemoryQueryDriver[MockModel]) *dummy.InMemoryQueryDriver[MockModel] {
			return imqd.WithCreate(func(_ *gin.Context, _ models.InternalValue) (models.InternalValue, error) {
				return models.InternalValue{}, errors.New("foo")
			})
		},
		wantStatus:          500,
		wantBodyContainsStr: "internal server error",
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

func TestCreateModelView(t *testing.T) {
	for _, tt := range createModelViewTests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			driverMod := tt.driverMod
			if driverMod == nil {
				driverMod = func(d *dummy.InMemoryQueryDriver[MockModel]) *dummy.InMemoryQueryDriver[MockModel] {
					return d
				}
			}
			_, r := gin.CreateTestContext(httptest.NewRecorder())
			r.POST("/foos", CreateModelViewSetFunc(
				IDFromQueryParamIDFunc,
				driverMod(queries.InMemory(MockModel{Foo: "bar"})),
				serializers.NewModelSerializer[MockModel](),
			))
			request, requestErr := http.NewRequest(http.MethodPost, "/foos", bytes.NewBufferString(tt.json))
			w := httptest.NewRecorder()

			// when
			r.ServeHTTP(w, request)

			// then
			a, readErr := io.ReadAll(w.Body)
			assert.NoError(t, requestErr)
			assert.NoError(t, readErr)
			if tt.wantBodyContainsStr != "" {
				assert.Contains(t, string(a), tt.wantBodyContainsStr)
			}
			if tt.wantBodyJSONEquals != nil {
				var responseJSON any
				responseJSONErr := json.Unmarshal(a, &responseJSON)
				assert.NoError(t, responseJSONErr)
				assert.Equal(t, tt.wantBodyJSONEquals, responseJSON)
			}
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
