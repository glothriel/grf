---
sidebar_position: 1
---

# Getting started

GRF is a library, that automatically generates REST APIs for GORM models using Gin. The simplest cases require merely few lines of code to generate a full REST resource with GET(list), GET(retrieve), POST(create), PUT(update), PATCH(update), DELETE(remove) methods and type validation. You can safely use GRF in your existing Gin application, it does not enforce any specific file layout or pattern.

## Full example

Here's a minimal example - it generates views supporting POST(create) and GET(list) methods, on `/people` path. The create action validates if name is present, otherwise throws validation error. On top of that any superflous fields are also reported as errors.

```go
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
```

Let's run such program and check how it works. First it should display an empty list:

```sh
$ go run main.go   
$ curl  http://localhost:8080/people -s | jq                            
[]
```

Let's add a person:

```sh
$ curl -s -X POST -d '{"name": "Andreas"}' http://localhost:8080/people | jq
{
  "id": 1,
  "name": "Andreas"
}
```

And another one:

```sh
$ curl -s -X POST -d '{"name": "Teresa"}' http://localhost:8080/people | jq
{
  "id": 2,
  "name": "Teresa"
}
```

Now let's check if the users were indeed created as expected:

```sh
$ curl  http://localhost:8080/people -s | jq                            
[
  {
    "id": 1,
    "name": "Andreas"
  },
  {
    "id": 2,
    "name": "Teresa"
  }
]
```

Now let's decompose all the parts of the listing to better understand what's going on.

## Models

```go
type Person struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"size:191;column:name"`
}
```

In order to generate the views, we need a model. GRF uses standard GORM models, but has some additional requirements regarding the fields used in the struct, you can read more of that in models section. Assuming, that your model uses standard golang types, everything should be fine. 

## Serializers

```go
serializers.NewValidatingSerializer[Person](
	serializers.NewModelSerializer[Person](),
	serializers.NewGoPlaygroundValidator[Person](
		map[string]any{
			"name": "required",
		},
	),
)
```

Serializers are used to translate the objects from the external API representation (what is incoming from JSON API) to internal representation (how it's represented in golang types) and backwards. Default `ModelSerializer` automatically includes all the fields from underlying model. `ValidatingSerializer` is a decorator placed on other Serializer implementation, that adds a validation layer. Currently we support [go-playground/validator](https://github.com/go-playground/validator) as the most popular validating library, but the validator interface is straightforward and you can creaate your own with 5 lines of code.



## Views

```go

views.NewListCreateModelView[Person]("/people", db.NewStaticResolver(gormDB)).WithSerializer(
	// serializer here
).Register(ginEngine)
```

Views package generates Gin views, exactly as its name suggests. During view creation you need to pass a database resolver - it can be used if you implement some kind of tenant separation, or just ignored and use gorm object directly, like here.
