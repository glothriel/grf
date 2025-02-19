package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/views"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var gormDBs sync.Map

func registerModel[Model any](
	prefix string, dialector gorm.Dialector, registeredModels ...any,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	gDB, ok := gormDBs.Load(dialector.Name())
	if !ok {
		theDB, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		gDB = any(theDB)
		gormDBs.Store(dialector.Name(), gDB)
	}

	gormDB, ok := gDB.(*gorm.DB)
	if !ok {
		panic("failed to cast database")
	}

	views.NewModelViewSet[Model](prefix, queries.GORM[Model](gormDB).WithOrderBy(fmt.Sprintf("%s ASC", "created_at"))).Register(router)

	var entity Model
	registeredModels = append(registeredModels, entity)
	if migrateErr := gormDB.AutoMigrate(registeredModels...); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
	return router
}

func newRequest(method, url string, body map[string]any) *http.Request {
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

type requestTestCase struct {
	name         string
	t            *testing.T
	req          *http.Request
	expectedCode int
	expectedJSON any
}

func (tc *requestTestCase) Req(r *http.Request) *requestTestCase {
	tc.req = r
	return tc
}

func (tc *requestTestCase) ExCode(c int) *requestTestCase {
	tc.expectedCode = c
	return tc
}

func (tc *requestTestCase) ExJson(j any) *requestTestCase {
	tc.expectedJSON = j
	return tc
}

func stripFields(v any, fields []string) any {
	switch val := v.(type) {
	case map[string]any:
		// Create a new map to avoid modifying the original
		result := make(map[string]any)
		for k, v := range val {
			if !contains(fields, k) {
				// Recursively strip fields from nested values
				result[k] = stripFields(v, fields)
			}
		}
		return result
	case []any:
		// Create a new slice to avoid modifying the original
		result := make([]any, len(val))
		for i, item := range val {
			// Recursively strip fields from slice elements
			result[i] = stripFields(item, fields)
		}
		return result
	default:
		return v
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (tc *requestTestCase) Run(router *gin.Engine) string {
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

func newRequestTestCase(t *testing.T, name string) *requestTestCase {
	return &requestTestCase{
		t:    t,
		name: name,
	}
}
