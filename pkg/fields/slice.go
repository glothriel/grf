package fields

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

// SliceModelField is a field that represents a slice of any type
type SliceModelField[T any, Model any] []T

// Update implements serializer.FieldUpdater interface, and decorates original field to
// correctly parse types of slice items
func (s SliceModelField[T, M]) Update(f *Field[M]) {
	previousValueFunc := f.InternalValueFunc
	var collectionItem T
	var m M
	var typeName string

	if !reflect.ValueOf(collectionItem).IsValid() {
		typeName = "any"
	} else {
		typeName = reflect.TypeOf(collectionItem).String()
	}
	structField, structFieldErr := StructAttributeByJSONTag(m, f.ItsName)
	if structFieldErr != nil {
		logrus.Fatalf(
			`%s.%s: %s`,
			reflect.TypeOf(m).String(),
			f.ItsName,
			structFieldErr.Error(),
		)
	}
	if structField.Tag.Get("gorm") == "" || !strings.Contains(structField.Tag.Get("gorm"), "type:text") {
		logrus.Fatalf(
			"%s.%s: SliceField should be used with `gorm:\"type:text\"` tag",
			reflect.TypeOf(m).String(),
			f.ItsName,
		)
	}

	f.InternalValueFunc = func(rawMap map[string]any, key string) (any, error) {
		rawValue, err := previousValueFunc(rawMap, key)
		if err != nil {
			return nil, err
		}
		rawValueSlice, ok := rawValue.([]any)
		if !ok {
			return nil, errors.New("Should be a collection")
		}
		var correctTypeSlice []T
		for i, v := range rawValueSlice {
			typedValue, ok := v.(T)
			if !ok {
				return nil, fmt.Errorf("%s[%d] is not a valid %s", key, i, typeName)
			}
			correctTypeSlice = append(correctTypeSlice, typedValue)
		}
		return correctTypeSlice, nil
	}
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (s *SliceModelField[T, M]) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := SliceModelField[T, M]{}
	err := json.Unmarshal(bytes, &result)
	*s = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (s SliceModelField[T, M]) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}
