---
sidebar_position: 1
---

# Getting Started

GRF is a library that automatically generates REST handlers for Gin. The simplest cases require only a few lines of code to generate a full REST resource with GET(list), GET(retrieve), POST(create), PUT(update), PATCH(update), DELETE(remove) methods, and type validation. You can safely use GRF in your existing Gin application as it does not enforce any specific file layout or pattern. For the full experience, you should use GORM as your ORM, but you can include your own QueryDriver implementation if you use something else.

## Simple Example with Full REST API

Let's build a simple application that consists of:

1. A model that maps to a SQL table.
2. A default view that creates REST actions allowing interaction with the model.
3. Customizations that add validation and additional logic to the application.

Let's start with the minimal example:

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/views"
)

type Person struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"size:191;column:name"`
}

func main() {
	ginEngine := gin.Default()
	queryDriver := queries.InMemory[Person]()
	views.NewModelViewSet[Person]("/people", queryDriver).Register(ginEngine)
	log.Fatal(ginEngine.Run(":8080"))
}
```

Let's run such a program and check how it works. First, it should display an empty list:

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

Now, let's decompose all the parts of the listing to better understand what's going on.

### Models

```go
type Person struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
```

In order to generate the views, we need a model. GRF requires you to have an ID field in the representation; other requirements can differ between QueryDrivers. This example uses the `InMemory` query driver, but if you're using GORM, you should include relevant tags here, like `gorm:"primaryKey"`.

### ViewSets

```go
views.NewModelViewSet[Person]("/people", queryDriver).Register(ginEngine)
```

ViewSets generate multiple views at once and operate on actions, just like Django Rest Framework does. There are five of them:

- List (GET `/<path/`)
- Create (POST `/<path/`)
- Read (GET `/<path/<id>`)
- Update (PUT|PATCH `/<path/<id>`)
- Delete (DELETE `/<path/<id>`)

When you call `Register` on a ViewSet, it registers views for all the actions it was configured for. The `NewModelViewSet` function automatically configures all actions. If you'd like to customize the actions exposed by your ViewSet, you can read more in the [views](/docs/views) section. If you'd like to read more about query drivers, you can do this in the [query drivers](/docs/query-drivers) section.

## Changing How the Fields Are Interpreted

### Serializers

Currently, our program is not very usable. For example, there is no validation, and you can add a person with an empty name. A mechanism that translates JSON input to models and vice-versa in GRF is called a serializer. If you are familiar with Django Rest Framework's serializers, it's a really similar concept. The `NewModelViewset` automatically configured an instance of `ModelSerializer` for us. It's a special serializer that scans the model's fields and tries to automatically set up the correct translation from and to JSON, using its knowledge of standard library types (like `int` or `string`), interfaces (like `encoding.TextMarshaler`), and internally registered custom types. Let's replace it with a wrapper that will also provide a validation layer, using the excellent `go-playground/validator` library.

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
```

Now, adding people with empty names will be impossible. If the bundled `goplayground/validate` is insufficient for you, you can provide your own implementation of the `serializers.Validator` interface.