package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/serializers"
	"github.com/glothriel/gin-rest-framework/pkg/types"
	"github.com/glothriel/gin-rest-framework/pkg/views"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	models.BaseModel

	Name        string          `json:"name" gorm:"size:191;column:name" validate:"required"`
	Description string          `json:"description" gorm:"type:text;column:description" validate:"required"`
	Price       decimal.Decimal `json:"price" gorm:"type:decimal(19,4)" validate:"required"`
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "products.db", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()
	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	mapper := getTypeMapper()

	if migrateErr := db.AutoMigrate(&Product{}); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
	serializer := serializers.NewValidatingSerializer[Product](
		serializers.NewModelSerializer[Product](mapper),
	).WithValidator(
		&serializers.GoPlaygroundValidator[Product]{},
	)

	views.NewListCreateModelView[Product]("/products", db).WithFieldTypeMapper(mapper).WithSerializer(
		serializer,
	).WithListSerializer(
		serializers.NewModelSerializer[Product](mapper).
			WithExistingFields([]string{"id", "name"}),
	).WithFilter(
		func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
			if ctx.Query("name") != "" {
				return db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("name")))
			}
			return db
		},
	).Register(router)

	views.NewRetrieveUpdateDeleteModelView[Product]("/products/:id", db).WithFieldTypeMapper(mapper).WithSerializer(
		serializer,
	).Register(router)

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}

func getTypeMapper() *types.FieldTypeMapper {
	mapper := types.DefaultFieldTypeMapper()
	mapper.Register("decimal.Decimal", types.FieldType{
		InternalToResponse: func(v interface{}) (interface{}, error) {
			decimalV, ok := v.(decimal.Decimal)
			if ok {
				return decimalV.String(), nil
			}
			stringV, ok := v.(string)
			if ok {
				return stringV, nil
			}
			return nil, fmt.Errorf("Expected %s to be a decimal or a string, got %T", v, v)
		},
		RequestToInternal: func(v interface{}) (interface{}, error) {
			decimalStr, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("Expected %s to be a string", v)
			}
			return decimal.NewFromString(decimalStr)
		},
	})
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
	return mapper
}
