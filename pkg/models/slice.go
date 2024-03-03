package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// SliceField is a field that represents a slice of any type
type SliceField[T any] []T

func (s *SliceField[T]) FromRepresentation(rawValue any) error {
	rawValueSlice, ok := rawValue.([]any)
	if !ok {
		return errors.New("Is not a collection")
	}
	var correctTypeSlice []T
	for i, v := range rawValueSlice {
		typedValue, ok := v.(T)
		if !ok {
			var t T
			return fmt.Errorf("[%d] is not a valid %T", i, t)
		}
		correctTypeSlice = append(correctTypeSlice, typedValue)
	}
	*s = correctTypeSlice
	return nil
}

func (s SliceField[T]) ToRepresentation() (any, error) {
	return s, nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (s *SliceField[T]) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to parse the value from database: is not bytes:", value))
	}

	result := SliceField[T]{}
	err := json.Unmarshal(bytes, &result)
	*s = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (s SliceField[T]) Value() (driver.Value, error) {
	return json.Marshal(s)
}
