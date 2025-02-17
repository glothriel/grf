package integration

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/shopspring/decimal"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		okBodies    []map[string]any
		errorBodies []map[string]any
		okResponses []map[string]any
		router      func() *gin.Engine
	}{
		{
			name:    "Bool type",
			baseURL: "/bool_field",
			okBodies: []map[string]any{
				{"value": true},
				{"value": false},
			},
			okResponses: []map[string]any{
				{"value": true},
				{"value": false},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": "hueble"},
				{"value": "true"},
				{"value": "True"},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value bool `json:"value" gorm:"column:value"`
				}]("/bool_field")
			},
		},
		{
			name:    "String type",
			baseURL: "/string_field",
			okBodies: []map[string]any{
				{"value": "hello world"},
				{"value": ""},
			},
			okResponses: []map[string]any{
				{"value": "hello world"},
				{"value": ""},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value string `json:"value" gorm:"column:value"`
				}]("/string_field")
			},
		},
		{
			name:    "Integer type",
			baseURL: "/int_field",
			okBodies: []map[string]any{
				{"value": 1337},
				{"value": -1337},
				{"value": 13.00},
			},
			okResponses: []map[string]any{
				{"value": 1337.0},
				{"value": -1337.0},
				{"value": 13.0},
			},
			errorBodies: []map[string]any{
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value int `json:"value" gorm:"column:value"`
				}]("/int_field")
			},
		},
		{
			name:    "Unsigned Integer type",
			baseURL: "/uint_field",
			okBodies: []map[string]any{
				{"value": 1337},
				{"value": 13.00},
			},
			okResponses: []map[string]any{
				{"value": 1337.0},
				{"value": 13.0},
			},
			errorBodies: []map[string]any{
				{"value": -1337},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value uint `json:"value" gorm:"column:value"`
				}]("/uint_field")
			},
		},
		{
			name:    "Float type",
			baseURL: "/float_field",
			okBodies: []map[string]any{
				{"value": 1.337},
				{"value": 1337},
			},
			okResponses: []map[string]any{
				{"value": 1.337},
				{"value": 1337.0},
			},
			errorBodies: []map[string]any{
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value float64 `json:"value" gorm:"column:value"`
				}]("/float_field")
			},
		},
		{
			name:    "String Slice type",
			baseURL: "/string_slice_field",
			okBodies: []map[string]any{
				{"value": []string{"hello world"}},
			},
			okResponses: []map[string]any{
				{"value": []any{"hello world"}},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{"1", 2}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value models.SliceField[string] `json:"value" gorm:"column:value;type:json"`
				}]("/string_slice_field")
			},
		},
		{
			name:    "Float Slice type",
			baseURL: "/float_slice_field",
			okBodies: []map[string]any{
				{"value": []float64{1.337}},
				{"value": []int{1, 2, 3}},
			},
			okResponses: []map[string]any{
				{"value": []any{1.337}},
				{"value": []any{1.0, 2.0, 3.0}},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []interface{}{"1", 2}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value models.SliceField[float64] `json:"value" gorm:"column:value;type:json"`
				}]("/float_slice_field")
			},
		},
		{
			name:    "time.Time type",
			baseURL: "/time",
			okBodies: []map[string]any{
				{"value": "2021-01-01T00:00:00Z"},
				{"value": "2021-01-01T00:00:00+00:00"},
			},
			okResponses: []map[string]any{
				{"value": "2021-01-01T00:00:00Z"},
				{"value": "2021-01-01T00:00:00Z"},
			},
			errorBodies: []map[string]any{
				{"value": "2021-01-01"},
				{"value": 2021},
				{"value": "01-01-2021T00:00:00Z"},
				{"value": "2021-13-01T00:00:00Z"},
				{"value": "2021-01-32T00:00:00Z"},
				{"value": "2021-01-01T24:00:00Z"},
				{"value": "2021-01-01T00:60:00Z"},
				{"value": "2021-01-01T00:00:60Z"},
				{"value": "2021-01-01T00:00"},
				{"value": "2021-01-01 00:00:00Z"},
				{"value": "2021-01-01T00:00:00"},
				{"value": "not-a-date"},
				{"value": ""},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value time.Time `json:"value" gorm:"column:value"`
				}]("/time")
			},
		},
		{
			name:    "Bool Slice type",
			baseURL: "/bool_slice_field",
			okBodies: []map[string]any{
				{"value": []bool{true}},
				{"value": []bool{false}},
			},
			okResponses: []map[string]any{
				{"value": []any{true}},
				{"value": []any{false}},
			},
			errorBodies: []map[string]any{
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{"1", 2}},
				{"value": []string{"true", "false"}},
				{"value": map[string]string{"foo": "bar"}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value models.SliceField[bool] `json:"value" gorm:"column:value;type:json"`
				}]("/bool_slice_field")
			},
		},
		{
			name:    "Any Slice type",
			baseURL: "/any_slice_field",
			okBodies: []map[string]any{
				{"value": []interface{}{true, "1", 2, 3.45}},
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{
					map[string]string{"foo": "bar"},
					map[string]string{"bar": "baz"},
					1,
					true,
				}},
				{"value": []string{"true", "false"}},
			},
			okResponses: []map[string]any{
				{"value": []interface{}{true, "1", 2.0, 3.45}},
				{"value": []interface{}{1.0, 2.0, 3.0}},
				{"value": []interface{}{
					any(map[string]any{"foo": "bar"}),
					any(map[string]any{"bar": "baz"}),
					1.0,
					true,
				}},
				{"value": []interface{}{"true", "false"}},
			},
			errorBodies: []map[string]any{
				{"value": map[string]any{"foo": "bar"}},
				{"value": true},
				{"value": false},
				{"value": "ads"},
				{"value": 1},
				{"value": 1.23},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value models.SliceField[any] `json:"value" gorm:"column:value;type:json"`
				}]("/any_slice_field")
			},
		},
		{
			name:    "Nullable Bool type",
			baseURL: "/null_bool_field",
			okBodies: []map[string]any{
				{"value": true},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": true},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value sql.NullBool `json:"value" gorm:"column:value"`
				}]("/null_bool_field")
			},
		},
		{
			name:    "Nullable String type",
			baseURL: "/null_string_field",
			okBodies: []map[string]any{
				{"value": "hello world"},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": "hello world"},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value sql.NullString `json:"value" gorm:"column:value"`
				}]("/null_string_field")
			},
		},
		{
			name:    "Decimal type",
			baseURL: "/decimal_field",
			okBodies: []map[string]any{
				{"value": "1.337"},
				{"value": "1337"},
				{"value": "13.37"},
				{"value": "0"},
			},
			okResponses: []map[string]any{
				{"value": "1.337"},
				{"value": "1337"},
				{"value": "13.37"},
				{"value": "0"},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": "hello world"},
				{"value": "1,337"},
				{"value": "1.3.37"},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value decimal.Decimal `json:"value" gorm:"column:value"`
				}]("/decimal_field")
			},
		},
		{
			name:    "Nullable Float type",
			baseURL: "/null_float64_field",
			okBodies: []map[string]any{
				{"value": 1.337},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": 1.337},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[struct {
					models.BaseModel
					Value sql.NullFloat64 `json:"value" gorm:"column:value"`
				}]("/null_float64_field")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if len(tt.okBodies) != len(tt.okResponses) {
				t.Fatalf("Test case %s: number of bodies and responses should be equal", tt.name)
			}

			for _, errorBody := range tt.errorBodies {
				NewAssertedReq(
					t,
					fmt.Sprintf("%s error: %v", tt.name, errorBody),
				).Req(
					NewRequest("POST", tt.baseURL, errorBody),
				).ExCode(
					http.StatusBadRequest,
				).Run(tt.router())
			}

			for i, okBody := range tt.okBodies {
				router := tt.router()
				resourceID := NewAssertedReq(
					t,
					fmt.Sprintf("%s create", tt.name),
				).Req(
					NewRequest("POST", tt.baseURL, okBody),
				).ExCode(
					http.StatusCreated,
				).ExJson(
					tt.okResponses[i],
				).Run(router)

				NewAssertedReq(
					t,
					tt.name,
				).Req(
					NewRequest("GET", tt.baseURL, nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					[]any{tt.okResponses[i]},
				).Run(router)

				NewAssertedReq(
					t,
					tt.name,
				).Req(
					NewRequest("GET", fmt.Sprintf("%s/%s", tt.baseURL, resourceID), nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					tt.okResponses[i],
				).Run(router)

				NewAssertedReq(
					t,
					tt.name,
				).Req(
					NewRequest("DELETE", fmt.Sprintf("%s/%s", tt.baseURL, resourceID), nil),
				).ExCode(
					http.StatusNoContent,
				).Run(router)

				NewAssertedReq(
					t,
					tt.name,
				).Req(
					NewRequest("GET", tt.baseURL, nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					[]any{},
				).Run(router)
			}

		})
	}
}
