package gormq

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
