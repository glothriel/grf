package fields

import (
	"fmt"
	"reflect"
)

func StructAttributeByJSONTag(theStruct interface{}, jsonTag string) (reflect.StructField, error) {
	structType := reflect.TypeOf(theStruct)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Tag.Get("json") == jsonTag {
			return field, nil
		}
	}
	return reflect.StructField{}, fmt.Errorf("no field with json tag %s", jsonTag)
}
