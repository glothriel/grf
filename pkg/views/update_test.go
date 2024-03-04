package views

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		name:               "Invalid Fields",
		json:               `{"bar": "baz"}`,
		wantStatus:         200,
		wantBodyJSONEquals: map[string]any{"id": float64(2), "foo": "bar"},
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

var enrichBodyWithIDTests = []struct {
	name           string
	isNumeric      bool
	idf            IDFunc
	body           map[string]any
	expectedResult map[string]any
	expectedError  error
}{
	{
		name:           "Numeric ID, ID in body matches ID in URL",
		isNumeric:      true,
		idf:            func(ctx *gin.Context) string { return "123" },
		body:           map[string]any{"id": float64(123), "foo": "bar"},
		expectedResult: map[string]any{"id": float64(123), "foo": "bar"},
		expectedError:  nil,
	},
	{
		name:           "Numeric ID, ID in body does not match ID in URL",
		isNumeric:      true,
		idf:            func(ctx *gin.Context) string { return "234" },
		body:           map[string]any{"id": float64(456), "foo": "bar"},
		expectedResult: nil,
		expectedError:  &serializers.ValidationError{FieldErrors: map[string][]string{"id": {"id in body does not match id in url"}}},
	},
	{
		name:           "Numeric ID, ID in body does not exist",
		isNumeric:      true,
		idf:            func(ctx *gin.Context) string { return "345" },
		body:           map[string]any{"foo": "bar"},
		expectedResult: map[string]any{"id": float64(345), "foo": "bar"},
		expectedError:  nil,
	},
	{
		name:           "ID is not numeric, ID in body matches ID in URL",
		isNumeric:      false,
		idf:            func(ctx *gin.Context) string { return "123" },
		body:           map[string]any{"id": "123", "foo": "bar"},
		expectedResult: map[string]any{"id": "123", "foo": "bar"},
		expectedError:  nil,
	},
	{
		name:           "ID is not numeric, ID in body does not match ID in URL",
		isNumeric:      false,
		idf:            func(ctx *gin.Context) string { return "234" },
		body:           map[string]any{"id": "456", "foo": "bar"},
		expectedResult: nil,
		expectedError:  &serializers.ValidationError{FieldErrors: map[string][]string{"id": {"id in body does not match id in url"}}},
	},
	{
		name:           "ID is not numeric, ID in body does not exist",
		isNumeric:      false,
		idf:            func(ctx *gin.Context) string { return "345" },
		body:           map[string]any{"foo": "bar"},
		expectedResult: map[string]any{"id": "345", "foo": "bar"},
		expectedError:  nil,
	},
	{
		name:           "ID is declared numeric, but is not numeric",
		isNumeric:      true,
		idf:            func(ctx *gin.Context) string { return "not a number" },
		body:           map[string]any{"id": float64(123), "foo": "bar"},
		expectedResult: nil,
		expectedError: &strconv.NumError{
			Func: "ParseFloat",
			Num:  "not a number",
			Err:  strconv.ErrSyntax,
		},
	},
}

func TestEnrichBodyWithID(t *testing.T) {
	for _, tt := range enrichBodyWithIDTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &gin.Context{}
			result, err := enrichBodyWithID[MockModel](ctx, tt.isNumeric, tt.idf, tt.body)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
