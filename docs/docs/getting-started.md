---
sidebar_position: 1
---

# Getting started

GRF is a library, that automatically generates REST handlers for Gin. The simplest cases require merely few lines of code to generate a full REST resource with GET(list), GET(retrieve), POST(create), PUT(update), PATCH(update), DELETE(remove) methods and type validation. You can safely use GRF in your existing Gin application, it does not enforce any specific file layout or pattern. For full experience you should use GORM as your ORM, but you can include your own QueryDriver implementation if you use something else.

## Full example

Here's a minimal example - it generates views supporting POST(create) and GET(list) methods, on `/people` path. It stores all the data in memory, so it's not very useful, but it's enough to show how the framework works:

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"
)

type Person struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func main() {
	ginEngine := gin.Default()
	views.NewListCreateModelView[Person]("/people", queries.InMemory()).WithSerializer(
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
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
```

In order to generate the views, we need a model. GRF requres you to have an `id` field in the representation, other requirements can differ between QueryDrivers. This example uses InMemory one, but if you'd use GORM, you should include relevant tags here, like `gorm:"primaryKey"`.

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

Serializers are used to translate the objects from the external API representation (what is incoming from JSON API) to internal representation (how it's represented in golang types) and backwards. Default `ModelSerializer` automatically includes all the fields from underlying model. `ValidatingSerializer` is a decorator placed on other Serializer implementation, that adds a validation layer. Currently we support [go-playground/validator](https://github.com/go-playground/validator) as the most popular validating library, but the validator interface is straightforward and you can create your own with 5 lines of code.

## Views

```go

views.NewListCreateModelView[Person]("/people", queries.InMemory()).WithSerializer(
	// serializer here
).Register(ginEngine)
```

Views package generates Gin views, exactly as its name suggests. During view creation you need to pass a QueryDriver, which is a storage layer of GRF. You can read more about query drivers in the [dedicated section](/docs/query-drivers).
