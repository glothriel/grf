package models

import (
	"reflect"
	"strings"
)

const tagID = "grf"

// TagIsRelation is a tag that indicates that the field is a relation.
const TagIsRelation = "relation"

// ParseTag parses the tag and returns a map of key-value pairs.
// Forma: `grf:"key1:value1;key2:value2"`
func ParseTag(f reflect.StructField) map[string]string {
	tag := f.Tag.Get(tagID)
	if tag == "" {
		return map[string]string{}
	}

	theMap := map[string]string{}

	pairs := strings.Split(tag, ";")

	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, ":", 2)
		key := strings.TrimSpace(kv[0])
		if key == "" {
			continue
		}
		if len(kv) == 1 {
			theMap[key] = ""
			continue
		}
		value := strings.TrimSpace(kv[1])
		theMap[key] = value
	}

	return theMap
}
