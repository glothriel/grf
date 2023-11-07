package views

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/stretchr/testify/assert"
)

var retrieveModelViewTests = []struct {
	name             string
	id               string
	wantStatus       int
	wantResponseBody map[string]interface{}
}{
	{
		name:             "Valid",
		id:               "1",
		wantStatus:       http.StatusOK,
		wantResponseBody: map[string]interface{}{"foo": "bar", "id": float64(1)},
	},
	{
		name:             "404",
		id:               "2",
		wantStatus:       http.StatusNotFound,
		wantResponseBody: map[string]interface{}{"message": "not found"},
	},
	// Add more test cases as needed
}

func TestRetrieveModelView(t *testing.T) {
	for _, tt := range retrieveModelViewTests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qd := queries.InMemory(MockModel{Foo: "bar"})
			serializer := serializers.NewModelSerializer[MockModel]()

			_, r := gin.CreateTestContext(httptest.NewRecorder())

			r.GET("/foos/:id", RetrieveModelViewSetFunc(IDFromQueryParamIDFunc, qd, serializer))

			retrieveRequest, retrieveRequestErr := http.NewRequest(http.MethodGet, "/foos/"+tt.id, nil)
			retrieveRecorder := httptest.NewRecorder()

			// when
			r.ServeHTTP(retrieveRecorder, retrieveRequest)

			// then
			assert.NoError(t, retrieveRequestErr)
			assert.Equal(t, tt.wantStatus, retrieveRecorder.Code)
			var actualResponseBody map[string]interface{}
			if err := json.Unmarshal(retrieveRecorder.Body.Bytes(), &actualResponseBody); err != nil {
				t.Errorf("Error parsing JSON response: %v", err)
			}
			assert.Equal(t, tt.wantResponseBody, actualResponseBody)
		})
	}
}
