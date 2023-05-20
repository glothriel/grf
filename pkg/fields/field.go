package fields

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/sirupsen/logrus"
)

type RepresentationFunc[Model any] func(*models.InternalValue[Model], string) (interface{}, error)
type InternalValueFunc func(map[string]interface{}, string) (interface{}, error)

type Field[Model any] struct {
	ItsName            string
	RepresentationFunc RepresentationFunc[Model]
	InternalValueFunc  InternalValueFunc
	FromDBFunc         InternalValueFunc
	Readable           bool
	Writable           bool
}

func (s *Field[Model]) Name() string {
	return s.ItsName
}

func (s *Field[Model]) ToRepresentation(intVal *models.InternalValue[Model]) (interface{}, error) {
	return s.RepresentationFunc(intVal, s.ItsName)
}

func (s *Field[Model]) ToInternalValue(reprModel map[string]interface{}) (interface{}, error) {
	return s.InternalValueFunc(reprModel, s.ItsName)
}

func (s *Field[Model]) FromDB(reprModel map[string]interface{}) (interface{}, error) {
	return s.FromDBFunc(reprModel, s.ItsName)
}

func (s *Field[Model]) ReadOnly() *Field[Model] {
	s.Readable = true
	s.Writable = false
	return s
}

func (s *Field[Model]) WriteOnly() *Field[Model] {
	s.Readable = false
	s.Writable = true
	return s
}

func (s *Field[Model]) ReadWrite() *Field[Model] {
	s.Readable = true
	s.Writable = true
	return s
}

func (s *Field[Model]) WithRepresentationFunc(f RepresentationFunc[Model]) *Field[Model] {
	s.RepresentationFunc = f
	return s
}

func (s *Field[Model]) WithInternalValueFunc(f InternalValueFunc) *Field[Model] {
	s.InternalValueFunc = f
	return s
}

func (s *Field[Model]) WithFromDBFunc(f InternalValueFunc) *Field[Model] {
	s.FromDBFunc = f
	return s
}

func NewField[Model any](name string) *Field[Model] {
	return &Field[Model]{
		ItsName: name,
		RepresentationFunc: func(intVal *models.InternalValue[Model], name string) (interface{}, error) {
			return intVal.Map[name], nil
		},
		FromDBFunc:        TrySQLScannerOrPassthrough[Model](),
		InternalValueFunc: InternalValuePassthrough(),
		Readable:          true,
		Writable:          true,
	}
}

func InternalValuePassthrough() InternalValueFunc {
	return func(reprModel map[string]interface{}, name string) (interface{}, error) {
		return reprModel[name], nil
	}
}

func TrySQLScannerOrPassthrough[Model any]() func(map[string]interface{}, string) (interface{}, error) {
	var entity Model
	jsonTagsToFieldNames := map[string]string{}
	for _, field := range reflect.VisibleFields(reflect.TypeOf(entity)) {
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			jsonTagsToFieldNames[jsonTag] = field.Name
		}
	}
	fieldBlueprints := map[string]reflect.Type{}
	for fieldName := range jsonTagsToFieldNames {
		ttype := reflect.TypeOf(
			reflect.ValueOf(entity).FieldByName(jsonTagsToFieldNames[fieldName]).Interface(),
		)
		fieldBlueprints[fieldName] = ttype
	}

	return func(reprModel map[string]interface{}, name string) (interface{}, error) {
		reflectedInstance := reflect.New(fieldBlueprints[name])

		var realFieldValue any
		if reflectedInstance.CanAddr() {
			realFieldValue = reflectedInstance.Addr().Interface()
		} else {
			realFieldValue = reflectedInstance.Interface()
		}

		scanner, ok := realFieldValue.(sql.Scanner)
		if !ok {
			logrus.Debugf("Field `%s` is not a sql.Scanner, returning value as is", name)
			return reprModel[name], nil
		}

		scanErr := scanner.Scan(reprModel[name])
		if scanErr != nil {
			return nil, fmt.Errorf("could not convert field from db `%s`: %s", name, scanErr)
		}
		return scanner, nil
	}
}
