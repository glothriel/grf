package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/db"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"
)

type Person struct {
	Id   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"size:191;column:name"`
}

func (p Person) ID() any {
	return p.Id
}

func main() {
	ginEngine := gin.Default()
	// gormDB, openErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	// if openErr != nil {
	// 	log.Fatalf("Failed to connect database: %s", openErr)
	// }
	// if migrateErr := gormDB.AutoMigrate(&Person{}); migrateErr != nil {
	// 	log.Fatalf("Error migrating database: %s", migrateErr)
	// }
	// database := gormdb.Gorm[Person](gormDB)
	database := db.Dummy(
		Person{Id: 1, Name: "John"},
		Person{Id: 2, Name: "Jane"},
		Person{Id: 3, Name: "Jack"},
	)
	views.NewListCreateModelView[Person]("/people", database).WithSerializer(
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
