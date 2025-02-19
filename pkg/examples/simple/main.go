package main

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/views"
)

type Person struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func main() {
	ginEngine := gin.Default()
	queryDriver := queries.InMemory[Person]()
	personViewSet := views.NewModelViewSet[Person](
		"/people", queryDriver).WithActions(views.ActionList)

	// Register the ViewSet with your Gin engine
	personViewSet.Register(ginEngine)

	// Start your Gin server
	ginEngine.Run(":8080")
}
