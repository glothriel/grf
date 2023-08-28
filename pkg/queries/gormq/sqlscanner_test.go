package gormq

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockModel struct {
	Field string `json:"field"`
}
type mockSQLScannerField map[string]any

// Scan scan value into Jsonb, implements sql.Scanner interface
func (s *mockSQLScannerField) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}

	result := make(map[string]any)
	err := json.Unmarshal(bytes, &result)
	*s = result
	return err
}

type mockSQLScannerModel struct {
	Field mockSQLScannerField `json:"field"`
}

func TestSQLScannerOrPassthroughWhenIsSqlScanner(t *testing.T) {
	// given
	fromDBFunc := SQLScannerOrPassthrough[mockSQLScannerModel]()

	// when
	value, valueErr := fromDBFunc(map[string]any{"field": []byte(`{"foo": "bar"}`)}, "field", nil)

	// then
	assert.Equal(t, mockSQLScannerField{"foo": "bar"}, value)
	assert.Nil(t, valueErr)
}

func TestSQLScannerOrPassthroughWhenIsSqlScannerScanError(t *testing.T) {
	// given
	fromDBFunc := SQLScannerOrPassthrough[mockSQLScannerModel]()

	// when
	value, valueErr := fromDBFunc(map[string]any{"field": []byte(`{"foo": "bar`)}, "field", nil)

	// then
	assert.Nil(t, value)
	assert.Error(t, valueErr)
}

func TestSQLScannerOrPassthroughWhenIsNotSqlScanner(t *testing.T) {
	// given
	fromDBFunc := SQLScannerOrPassthrough[mockModel]()

	// when
	value, valueErr := fromDBFunc(map[string]any{"field": "foo"}, "field", nil)

	// then
	assert.Equal(t, "foo", value)
	assert.Nil(t, valueErr)
}
