package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InternalRepresentation is a type that is used to hold model data. We don't use structs,
// because it would force heavy use of reflection, which is slow.

// Base contains common columns for all tables.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

type InternalValue[Model any] map[string]any

func (i InternalValue[Model]) AsModel() (Model, error) {
	var entity Model
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &entity,
		Squash:  true,
	})

	if decoderErr != nil {
		return entity, decoderErr
	}
	decodeErr := decoder.Decode(i)
	if decodeErr != nil {
		logrus.Debug(i)
		return entity, fmt.Errorf(
			"Failed to convert internal value to model `%T`: Mapstructure error: %w", entity, decodeErr,
		)
	}
	return entity, nil
}

// function that converts a map to a struct using reflection without mapstructure
// use json keys as struct field names
func MapToStruct[Model any](m map[string]any) (Model, error) {
	var entity Model
	jsonTagsToFieldNames := map[string]string{}
	t := reflect.TypeOf(entity)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			jsonTagsToFieldNames[jsonTag] = field.Name
		}
	}
	v := reflect.ValueOf(&entity).Elem()
	for k, val := range m {
		v.FieldByName(jsonTagsToFieldNames[k]).Set(reflect.ValueOf(val))
	}
	return entity, nil
}

// Uses reflect package to do a shallow translation of the model to a map wrapped in an InternalValue. We
// could use mapstructure, but it does a deep translation, which is not what we want.
func InternalValueFromModel[Model any](entity Model) (InternalValue[Model], error) {
	v := reflect.ValueOf(entity)
	out := make(InternalValue[Model])
	fields := reflect.VisibleFields(reflect.TypeOf(entity))
	for _, field := range fields {
		if !field.Anonymous {
			out[field.Tag.Get("json")] = v.FieldByName(field.Name).Interface()
		}
	}
	return out, nil
}
