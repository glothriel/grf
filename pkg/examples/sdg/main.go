package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/glothriel/gin-rest-framework/pkg/fields"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/serializers"
	"github.com/glothriel/gin-rest-framework/pkg/types"
	"github.com/glothriel/gin-rest-framework/pkg/views"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"reflect"
)

/*
{
	"enabled": true,
  "integration": "acceptance",
	"api_key": "SEmoku59bwgUagPHXyv3EmWQ",
	"url_filter": false,
	"url_contains": [
		"blog",
		"/categories/unicorn"
	]
}
*/

type SliceOfItems[T any] []T

func (s SliceOfItems[T]) Process(f *fields.Field[SDGConfig]) {
	previousValueFunc := f.InternalValueFunc
	var m T
	typeName := reflect.TypeOf(m).Name()

	f.InternalValueFunc = func(rawMap map[string]interface{}, key string) (interface{}, error) {
		previousValue, err := previousValueFunc(rawMap, key)
		if err != nil {
			return nil, err
		}
		previousValueType, ok := previousValue.([]interface{})
		if !ok {
			return nil, &serializers.ValidationError{map[string][]string{"url_contains": {fmt.Sprintf("Expected %s", typeName)}}}
		}
		for i, v := range previousValueType {
			if _, ok := v.(T); !ok {
				return nil, fmt.Errorf("url_contains[%d] is not a %s", i, typeName)
			}
		}
		return previousValueType, nil
	}
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *SliceOfItems[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := SliceOfItems[T]{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (j SliceOfItems[T]) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

//type SliceOfStrings []string
//
//// Scan scan value into Jsonb, implements sql.Scanner interface
//func (j *SliceOfStrings) Scan(value interface{}) error {
//	bytes, ok := value.([]byte)
//	if !ok {
//		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
//	}
//
//	result := SliceOfStrings{}
//	err := json.Unmarshal(bytes, &result)
//	*j = result
//	return err
//}
//
//// Value return json value, implement driver.Valuer interface
//func (j SliceOfStrings) Value() (driver.Value, error) {
//	if len(j) == 0 {
//		return nil, nil
//	}
//	return json.Marshal(j)
//}

/*
url_contains
"string" "string"

*/

type SDGConfig struct {
	models.BaseModel

	Enabled     bool   `json:"enabled" validate:"required"`
	Integration string `json:"integration" gorm:"type:text;column:integration" validate:"required"` // FIXME enum
	ApiKey      string `json:"api_key" gorm:"column:api_key" validate:"required"`
	// go-playground validator doesn't support bools - workaround is to remove required and set default value
	UrlFilter bool `json:"url_filter" gorm:"default:false"`
	// TODO validate is json an array of strings
	UrlContains SliceOfItems[string] `json:"url_contains" gorm:"type:text"` // FIXME array of strings
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "sdg.db", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()
	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	mapper := getTypeMapper()

	if migrateErr := db.AutoMigrate(&SDGConfig{}); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
	serializer := serializers.NewValidatingSerializer[SDGConfig](
		serializers.NewModelSerializer[SDGConfig](mapper)).WithValidator(
		&serializers.GoPlaygroundValidator[SDGConfig]{},
	)

	views.NewListCreateModelView[SDGConfig]("/sdg", db).WithFieldTypeMapper(
		mapper,
	).WithSerializer(
		serializer,
	).WithFilter(
		func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
			if ctx.Query("api_key") != "" {
				return db.Where("api_key LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("api_key")))
			}
			return db
		},
	).Register(router)

	views.NewRetrieveUpdateDeleteModelView[SDGConfig]("/sdg/:id", db).WithFieldTypeMapper(mapper).WithSerializer(
		serializer,
	).Register(router)

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}

func getTypeMapper() *types.FieldTypeMapper {
	mapper := types.DefaultFieldTypeMapper()
	mapper.Register("uuid.UUID", types.FieldType{
		InternalToResponse: func(v interface{}) (interface{}, error) {
			return v, nil
		},
		RequestToInternal: func(v interface{}) (interface{}, error) {
			theUUID, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("Expected %s to be a string", v)
			}
			return uuid.Parse(theUUID)
		},
	})
	mapper.Register("main.SliceOfItems[string]", types.FieldType{
		InternalToResponse: types.ConvertPassThrough,
		RequestToInternal:  types.ConvertPassThrough,
	})
	return mapper
}
