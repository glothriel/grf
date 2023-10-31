package main

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	models.BaseModel

	Name        string          `json:"name" gorm:"size:191;column:name"`
	Description string          `json:"description" gorm:"type:text;column:description"`
	Price       decimal.Decimal `json:"price" gorm:"type:decimal(19,4)"`
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "products.db", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()

	gormDB, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if migrateErr := gormDB.AutoMigrate(&Product{}); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}

	v := views.NewModelViewSet[Product](
		"/products",
		queries.GORM[Product](gormDB).WithFilter(
			func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				if ctx.Query("name") != "" {
					return db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("name")))
				}
				return db
			},
		).WithOrderBy("name ASC"),
	).WithSerializer(
		serializers.NewValidatingSerializer[Product](
			serializers.NewModelSerializer[Product](),
			serializers.NewGoPlaygroundValidator[Product](
				map[string]any{
					"name":        "required",
					"description": "required",
				},
			),
		),
	).WithListSerializer(
		serializers.NewModelSerializer[Product]().
			WithModelFields([]string{"id", "name"}),
	)

	fmt.Println(reflect.TypeOf(v.CreateAction.Serializer).String())

	v.Register(router)

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}
