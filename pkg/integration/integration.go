package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/views"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func registerModel[Model any](
	prefix string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	views.NewModelViewSet[Model](prefix, queries.GORM[Model](gormDB).WithOrderBy(fmt.Sprintf("%s ASC", "created_at"))).Register(router)

	var entity Model
	if migrateErr := gormDB.AutoMigrate(&entity); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
	return router
}

func NewRequest(method, url string, body map[string]any) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		marshalledBody, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			panic("failed to marshal body")
		}
		req.Body = io.NopCloser(bytes.NewReader(marshalledBody))
	}
	return req
}

type TestCase struct {
	name         string
	t            *testing.T
	req          *http.Request
	expectedCode int
	expectedJSON any
}

func (tc *TestCase) Req(r *http.Request) *TestCase {
	tc.req = r
	return tc
}

func (tc *TestCase) ExCode(c int) *TestCase {
	tc.expectedCode = c
	return tc
}

func (tc *TestCase) ExJson(j any) *TestCase {
	tc.expectedJSON = j
	return tc
}

func stripFields(v any, fields []string) any {
	vMap, ok := v.(map[string]any)
	if ok {
		for _, field := range fields {
			delete(vMap, field)
		}
		return vMap
	}
	vSlice, ok := v.([]any)
	if ok {
		for i, item := range vSlice {
			vSlice[i] = stripFields(item, fields)
		}
		return vSlice
	}
	return v
}

func (tc *TestCase) Run(router *gin.Engine) string {
	var theID string
	w := httptest.NewRecorder()
	router.ServeHTTP(w, tc.req)

	tc.t.Run(tc.name, func(t *testing.T) {
		require.Equal(t, tc.expectedCode, w.Code)
		if tc.expectedJSON != nil {
			var responseJSON any
			json.Unmarshal(w.Body.Bytes(), &responseJSON)
			rAsMap, ok := responseJSON.(map[string]any)
			if ok {
				id, ok := rAsMap["id"]
				if ok {
					theID = id.(string)
				}
			}
			require.Equal(t, tc.expectedJSON, stripFields(responseJSON, []string{"id", "created_at", "updated_at"}))

		}
	})

	return theID
}

func NewAssertedReq(t *testing.T, name string) *TestCase {
	return &TestCase{
		t:    t,
		name: name,
	}
}
