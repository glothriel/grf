package views

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

type anotherMockModel struct {
	ID    uint    `json:"id"`
	Price float64 `json:"price"`
	Name  string  `json:"name"`
}

var nameOnlySerializer = serializers.NewModelSerializer[anotherMockModel]().WithModelFields([]string{"name"})

type quickReqParams struct {
	method string
	path   string
	body   func() io.Reader
}

func noBody() io.Reader {
	return nil
}

func strBody(s string) func() io.Reader {
	return func() io.Reader {
		return bytes.NewBufferString(s)
	}
}

func quickReq(r *gin.Engine, params quickReqParams) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(params.method, params.path, params.body())
	r.ServeHTTP(w, req)
	return w

}

type viewsetTestCase struct {
	name   string
	params quickReqParams
}

var caseList viewsetTestCase = viewsetTestCase{
	name: "GET request to collection endpoint",
	params: quickReqParams{
		method: "GET",
		path:   "/mocks",
		body:   noBody,
	},
}

var caseRetrieve viewsetTestCase = viewsetTestCase{
	name: "GET request to single item endpoint",
	params: quickReqParams{
		method: "GET",
		path:   "/mocks/1",
		body:   noBody,
	},
}

var caseCreate viewsetTestCase = viewsetTestCase{
	name: "POST request to collection endpoint",
	params: quickReqParams{
		method: "POST",
		path:   "/mocks",
		body:   strBody(`{"price": 1.0,"name": "Canned Beans"}`),
	},
}

var caseUpdate viewsetTestCase = viewsetTestCase{
	name: "PUT request to single item endpoint",
	params: quickReqParams{
		method: "PUT",
		path:   "/mocks/1",
		body:   strBody(`{"price": 2.0,"name": "Canned Beans"}`),
	},
}

var caseDestroy viewsetTestCase = viewsetTestCase{
	name: "DELETE request to single item endpoint",
	params: quickReqParams{
		method: "DELETE",
		path:   "/mocks/1",
		body:   noBody,
	},
}

func TestEmptyViewsetRespondsWithMethodNotFound(t *testing.T) {
	viewset := NewViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel]())
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	for _, tc := range []viewsetTestCase{
		caseCreate,
		caseRetrieve,
		caseList,
		caseUpdate,
		caseDestroy,
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := quickReq(r, tc.params)
			assert.Equal(t, 405, w.Code)
		})
	}
}
func TestViewsetWhenOnlyListActionRegisteredAllOthersReturn405(t *testing.T) {
	viewset := NewViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel]()).WithActions(ActionList)
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	for _, tt := range []viewsetTestCase{
		caseCreate,
		caseRetrieve,
		caseUpdate,
		caseDestroy,
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := quickReq(r, tt.params)
			assert.Equal(t, 405, w.Code)
		})
	}

	for _, tt := range []viewsetTestCase{
		caseList,
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := quickReq(r, tt.params)
			assert.Equal(t, 200, w.Code)
		})
	}
}

func checkAllActionsNoErrors(t *testing.T, v *ViewSet[anotherMockModel]) {
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	v.Register(r)
	for _, tt := range []viewsetTestCase{
		caseList,
		caseCreate,
		caseRetrieve,
		caseUpdate,
		caseDestroy,
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := quickReq(r, tt.params)
			assert.Less(t, w.Code, 299)
		})
	}
}

func TestNewModelViewset(t *testing.T) {
	checkAllActionsNoErrors(
		t, NewModelViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel]()),
	)
}
func TestNewViewSetAllActions(t *testing.T) {
	checkAllActionsNoErrors(
		t, NewViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel]()).WithActions(
			ActionList, ActionCreate, ActionRetrieve, ActionUpdate, ActionDestroy,
		),
	)
}

func TestCustomSerializer(t *testing.T) {
	// given
	viewset := NewModelViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel](
		anotherMockModel{Price: 1.0, Name: "Canned Beans"},
	)).WithSerializer(nameOnlySerializer)
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	// when
	w := quickReq(r, caseList.params)

	// then
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `[{"name":"Canned Beans"}]`, w.Body.String())
}

func TestCustomRetrieveSerializer(t *testing.T) {
	// given
	viewset := NewModelViewSet[anotherMockModel]("/mocks", queries.InMemory[anotherMockModel](
		anotherMockModel{Price: 1.0, Name: "Canned Beans"},
	)).WithRetrieveSerializer(nameOnlySerializer)
	_, r := gin.CreateTestContext(httptest.NewRecorder())
	viewset.Register(r)

	// when
	ls := quickReq(r, caseList.params)
	rt := quickReq(r, caseRetrieve.params)

	// then
	assert.Equal(t, 200, ls.Code)
	assert.Equal(t, 200, rt.Code)
	assert.Equal(t, `[{"id":1,"name":"Canned Beans","price":1}]`, ls.Body.String())
	assert.Equal(t, `{"name":"Canned Beans"}`, rt.Body.String())
}
