package fields

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/sirupsen/logrus"
)

type RepresentationFunc func(models.InternalValue, string, *gin.Context) (any, error)
type InternalValueFunc func(map[string]any, string, *gin.Context) (any, error)

type Field[Model any] struct {
	name               string
	representationFunc RepresentationFunc
	internalValueFunc  InternalValueFunc
	fromDBFunc         InternalValueFunc

	Readable bool
	Writable bool
}

func (s *Field[Model]) Name() string {
	return s.name
}

func (s *Field[Model]) ToRepresentation(intVal models.InternalValue, ctx *gin.Context) (any, error) {
	return s.representationFunc(intVal, s.name, ctx)
}

func (s *Field[Model]) ToInternalValue(reprModel map[string]any, ctx *gin.Context) (any, error) {
	return s.internalValueFunc(reprModel, s.name, ctx)
}

func (s *Field[Model]) FromDB(reprModel map[string]any, ctx *gin.Context) (any, error) {
	return s.fromDBFunc(reprModel, s.name, ctx)
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

func (s *Field[Model]) WithRepresentationFunc(f RepresentationFunc) *Field[Model] {
	s.representationFunc = f
	return s
}

func (s *Field[Model]) WithInternalValueFunc(f InternalValueFunc) *Field[Model] {
	s.internalValueFunc = f
	return s
}

func (s *Field[Model]) WithFromDBFunc(f InternalValueFunc) *Field[Model] {
	s.fromDBFunc = f
	return s
}

func NewField[Model any](name string) *Field[Model] {
	return &Field[Model]{
		name: name,
		representationFunc: func(intVal models.InternalValue, name string, ctx *gin.Context) (any, error) {
			return intVal[name], nil
		},
		fromDBFunc: SQLScannerOrPassthrough[Model](),
		internalValueFunc: func(reprModel map[string]any, name string, ctx *gin.Context) (any, error) {
			return reprModel[name], nil
		},
		Readable: true,
		Writable: true,
	}
}

func SQLScannerOrPassthrough[Model any]() func(map[string]any, string, *gin.Context) (any, error) {
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

	return func(reprModel map[string]any, name string, ctx *gin.Context) (any, error) {
		reflectedInstance := reflect.New(fieldBlueprints[name]).Interface()

		scanner, ok := reflectedInstance.(sql.Scanner)
		if !ok {
			logrus.Debugf("Field `%s` is not a sql.Scanner, returning value as is", name)
			return reprModel[name], nil
		}

		scanErr := scanner.Scan(reprModel[name])
		if scanErr != nil {
			return nil, fmt.Errorf("could not convert field from db `%s`: %s", name, scanErr)
		}

		return reflect.ValueOf(scanner).Elem().Interface(), nil
	}
}
