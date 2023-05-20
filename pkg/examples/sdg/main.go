package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/gin-rest-framework/pkg/fields"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/serializers"
	"github.com/glothriel/gin-rest-framework/pkg/views"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Foo struct {
	Bla string
	Goo int
}

type SDGConfig struct {
	models.BaseModel

	Enabled     bool   `json:"enabled" validate:"required"`
	Integration string `json:"integration" gorm:"type:text;column:integration" validate:"required"` // FIXME enum
	ApiKey      string `json:"api_key" gorm:"column:api_key" validate:"required"`
	// go-playground validator doesn't support bools - workaround is to remove required and set default value
	UrlFilter bool `json:"url_filter" gorm:"default:false"`
	// TODO validate is json an array of strings
	UrlContains fields.SliceField[int, SDGConfig] `json:"url_contains" gorm:"type:text"` // FIXME array of strings
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

	if migrateErr := db.AutoMigrate(&SDGConfig{}); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
	serializer := serializers.NewValidatingSerializer[SDGConfig](
		serializers.NewModelSerializer[SDGConfig](nil)).WithValidator(
		&serializers.GoPlaygroundValidator[SDGConfig]{},
	)

	views.NewListCreateModelView[SDGConfig]("/sdg", db).WithSerializer(
		serializer,
	).WithFilter(
		func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
			if ctx.Query("api_key") != "" {
				return db.Where("api_key LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("api_key")))
			}
			return db
		},
	).Register(router)

	views.NewRetrieveUpdateDeleteModelView[SDGConfig]("/sdg/:id", db).WithSerializer(
		serializer,
	).Register(router)

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}
