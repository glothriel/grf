package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"
)

type Person struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"size:191;column:name"`
}

func main() {
	ginEngine := gin.Default()
	queryDriver := queries.InMemory[Person]()
	serializer := serializers.NewValidatingSerializer[Person](
		serializers.NewModelSerializer[Person](),
		serializers.NewGoPlaygroundValidator[Person](
			map[string]any{
				"name": "required",
			},
		),
	)
	views.NewModelViewSet[Person]("/people", queryDriver).WithSerializer(serializer).Register(ginEngine)
	log.Fatal(ginEngine.Run(":8080"))
}
