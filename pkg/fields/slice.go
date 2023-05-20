package fields

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
)

type SliceField[T any, Model any] []T

// integers need to be defined as SliceOfItems[float64] because of the way that json.Unmarshal works
// this doesn't work for maps e.g SliceOfItems[map[string]string]
func (s SliceField[T, M]) Update(f *Field[M]) {
	previousValueFunc := f.InternalValueFunc
	var collectionItem T
	var m M
	typeName := reflect.TypeOf(collectionItem).String()
	if typeName == "int" {
		logrus.Fatalf(
			`%s.%s: SliceField does not support int, because JSON "number" type. Use float64 instead`,
			reflect.TypeOf(m).String(),
			f.ItsName,
		)
	}

	f.InternalValueFunc = func(rawMap map[string]interface{}, key string) (interface{}, error) {
		rawValue, err := previousValueFunc(rawMap, key)
		if err != nil {
			return nil, err
		}
		rawValueSlice, ok := rawValue.([]interface{})
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
func (s *SliceField[T, M]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := SliceField[T, M]{}
	err := json.Unmarshal(bytes, &result)
	*s = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (s SliceField[T, M]) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}
