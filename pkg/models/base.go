package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// InternalRepresentation is a type that is used to hold model data. We don't use structs,
// because it would force heavy use of reflection, which is slow.

// BaseModel contains common columns for all tables.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.New()
	return nil
}

// InternalValue is a map that holds model data. It is used to avoid heavy use of reflection
// during read operations.
type InternalValue map[string]any

// AsInternalValue Uses reflect package to do a shallow translation of the model to a map
// wrapped in an InternalValue. We could use mapstructure, but it does a deep translation,
// which is not what we want.
func AsInternalValue[Model any](entity Model) InternalValue {
	v := reflect.ValueOf(entity)
	out := make(InternalValue)
	fields := reflect.VisibleFields(reflect.TypeOf(entity))
	for _, field := range fields {
		if !field.Anonymous {
			out[field.Tag.Get("json")] = v.FieldByName(field.Name).Interface()
		}
	}
	return out
}

// AsModel decodes an InternalValue to a model. It uses mapstructure package to do so.
func AsModel[Model any](i InternalValue) (Model, error) {
	var entity Model
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &entity,
		Squash:  true,
	})

	if decoderErr != nil {
		return entity, decoderErr
	}
	if decodeErr := decoder.Decode(i); decodeErr != nil {
		return entity, fmt.Errorf(
			"Failed to convert internal value to model `%T`: Mapstructure error: %w", entity, decodeErr,
		)
	}
	return entity, nil
}
