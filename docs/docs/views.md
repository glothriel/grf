# Views

# Using ViewSets in GRF

ViewSets in GRF (Gin REST Framework) simplify the process of creating RESTful APIs by providing a structured way to define and manage CRUD (Create, Read, Update, Delete) operations for your data models. This documentation will guide you through the usage of ViewSets in GRF.

## Prerequisites

Before you start using ViewSets, make sure you have the following prerequisites in place:

- Go programming environment set up.
- GRF and Gin installed in your project.

## Creating a Basic ViewSet

Let's start by creating a basic ViewSet for a data model. In this example, we'll assume you have a `Person` model.

```go
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
	personViewSet := views.NewViewSet("/people", queryDriver).WithActions(views.ActionsList)

	// Register the ViewSet with your Gin engine
	personViewSet.Register(ginEngine)

	// Start your Gin server
	ginEngine.Run(":8080")
}
```

In this example, we create a basic ViewSet for the `Person` model, and we register it with the Gin engine. The viewset will only respond to List action.

## Configuring Actions

ViewSets provide the following actions:

- List (GET `/people`)
- Create (POST `/people`)
- Retrieve (GET `/people/:id`)
- Update (PUT/PATCH `/people/:id`)
- Destroy (DELETE `/people/:id`)

You can customize which actions are available by using the `WithActions` method:

```go
personViewSet.WithActions(views.ActionList, views.ActionCreate).Register(ginEngine)
```

In this example, we configure the ViewSet to only include the List and Create actions.

## Customizing Serializers

Serializers are responsible for translating JSON input to models and vice versa. You can customize the default serializer (`serializers.NewModelSerializer`, including all the fields) for the ViewSet or individual actions:

```go
serializer := MyCustomSerializer()

// custom serializer for the entire ViewSet
personViewSet.WithSerializer(serializer)

// custom serializer only for list action
personViewSet.WithListSerializer(serializer)
```

## Adding side effects

:::info

Side effects are only available for actions, that mutate the state of QueryDriver (update/create/destroy). If you'd like to add side effect to list or retrieve action, please use Serializer API.

:::

GRF allows adding side effects to your viewset actions. 

```go
personViewSet.OnCreate(customCreateLogic)
personViewSet.OnUpdate(customUpdateLogic)
personViewSet.OnDestroy(customDestroyLogic)
```

In this example, we add custom logic for the Create, Update, and Destroy operations.

## Registering the ViewSet

After configuring your ViewSet and Gin engine, make sure to call the `Register` method to register the ViewSet's routes:

```go
personViewSet.Register(ginEngine)
```

Now, your Gin server is ready to handle RESTful API requests for the `Person` model.

## Conclusion

ViewSets in GRF simplify the creation of RESTful APIs by providing a structured way to define and manage CRUD operations. With ViewSets, you can quickly set up endpoints for your data models and focus on customizing the behavior as needed.
