package views

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/stretchr/testify/assert"
)

type AnotherMockModel struct {
	ID    uint    `json:"id"`
	Price float64 `json:"price"`
	Name  string  `json:"name"`
}

type requestParams struct {
	method string
	path   string
	body   io.Reader
}

func quickReq(r *gin.Engine, params requestParams) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(params.method, params.path, params.body)
	r.ServeHTTP(w, req)
	return w

}

var testCases = []struct {
	name   string
	params requestParams
}{
	{
		name: "GET request to collection endpoint",
		params: requestParams{
			method: "GET",
			path:   "/mocks",
		},
	},
	{
		name: "GET request to single item endpoint",
		params: requestParams{
			method: "GET",
			path:   "/mocks/1",
		},
	},
	{
		name: "DELETE request to single item endpoint",
		params: requestParams{
			method: "DELETE",
			path:   "/mocks/1",
		},
	},
	{
		name: "POST request to collection endpoint",
		params: requestParams{
			method: "POST",
			path:   "/mocks",
			body:   bytes.NewBufferString(`{"price": 1.0}`),
		},
	},
	{
		name: "PUT request to single item endpoint",
		params: requestParams{
			method: "PUT",
			path:   "/mocks/1",
			body:   bytes.NewBufferString(`{"price": 1.0, "name": "bar"}`),
		},
	},
}

func TestEmptyViewsetRespondsWithMethodNotFound(t *testing.T) {
	viewset := NewViewSet[AnotherMockModel]("/mocks", queries.InMemory[AnotherMockModel]())
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := quickReq(r, tc.params)
			assert.Equal(t, 405, w.Code)
		})
	}
}
func TestViewsetOnlyListActionRegistered(t *testing.T) {
	viewset := NewViewSet[AnotherMockModel]("/mocks", queries.InMemory[AnotherMockModel]()).WithActions(ActionList)
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	// Tests for actions that should return 405
	nonListTests := []struct {
		name   string
		params requestParams
	}{
		{
			name: "GET single item should return 405",
			params: requestParams{
				method: "GET",
				path:   "/mocks/1",
			},
		},
		{
			name: "DELETE item should return 405",
			params: requestParams{
				method: "DELETE",
				path:   "/mocks/1",
			},
		},
		{
			name: "POST to collection should return 405",
			params: requestParams{
				method: "POST",
				path:   "/mocks",
				body:   bytes.NewBufferString("{}"),
			},
		},
		{
			name: "PUT single item should return 405",
			params: requestParams{
				method: "PUT",
				path:   "/mocks/1",
				body:   bytes.NewBufferString(`{"price": 1.0}`),
			},
		},
	}

	for _, tc := range nonListTests {
		t.Run(tc.name, func(t *testing.T) {
			w := quickReq(r, tc.params)
			assert.Equal(t, 405, w.Code)
		})
	}

	// Test for LIST action which should return 200
	t.Run("GET collection should return 200", func(t *testing.T) {
		w := quickReq(r, requestParams{
			method: "GET",
			path:   "/mocks",
		})
		assert.Equal(t, 200, w.Code)
	})
}
