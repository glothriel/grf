package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Person struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"size:191;column:name"`
}

func main() {
	ginEngine := gin.Default()
	gormDB, openErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if openErr != nil {
		log.Fatalf("Failed to connect database: %s", openErr)
	}
	if migrateErr := gormDB.AutoMigrate(&Person{}); migrateErr != nil {
		log.Fatalf("Error migrating database: %s", migrateErr)
	}
	views.NewListCreateModelView[Person]("/people", db.NewStaticResolver(gormDB)).WithSerializer(
		serializers.NewValidatingSerializer[Person](
			serializers.NewModelSerializer[Person](),
			serializers.NewGoPlaygroundValidator[Person](
				map[string]any{
					"name": "required",
				},
			),
		),
	).Register(ginEngine)
	log.Fatal(ginEngine.Run(":8080"))
}
