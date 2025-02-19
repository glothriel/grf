# Serializers

Serializers in GRF are responsible for translation of data between the database and the API. They control which fields are accepted in the payload, which are served in the response, and how they are formatted. The best way to understand how serializers work is to see them in action.

```go
type Serializer interface {
    // This function is called when transforming the request payload into an internal value
    // that will be later passed to QueryDriver (for example GORM) for processing.
	ToInternalValue(map[string]any, *gin.Context) (models.InternalValue, error)
    
    // This function transforms data obtained from the QueryDriver into a response payload, that
    // will be then marshalled into JSON and sent to the client.
	ToRepresentation(models.InternalValue, *gin.Context) (Representation, error)
}
```

## ModelSerializer

`ModelSerializer` is the workhorse of GRF. It can be created with `serializers.NewModelSerializer[Model]()`. By default such serializer will include all struct fields (there's also `NewEmptyModelSerializer[Model]()` that will not include any fields).


### Using existing fields in ModelSerializer

You can select which existing fields should be included in the response by using the `WithModelFields` method.

```go
serializer := serializers.NewModelSerializer[Model]().WithModelFields("field1", "field2")
```

Please note, that the fields here are identified by their JSON tags, not the struct field names.

### Adding completely new fields

For adding fields, that are not present in the model struct or you didn't want to include in the `WithModelFields` method, you can use the `WithField` method. More about fields can be found in the [Fields](#fields) section.

```go
serializer := serializers.NewModelSerializer[Model]().WithNewField(
    fields.NewField("color").ReadOnly().WithRepresentationFunc(
        func(models.InternalValue, string, *gin.Context) (any, error) {
            return "blue", nil
        }
    )
)
```

### Customizing existing fields

You can also customize existing fields by using the `WithField` method.

```go

serializer := serializers.NewModelSerializer[Model]().WithField(
    "color",
    func(oldField fields.Field){
        return oldField.WithRepresentationFunc(
            func(models.InternalValue, string, *gin.Context) (any, error) {
                return "pink", nil
            }
        )
    }
)
```

## Fields

Fields are used by ModelSerializers to transform data between the database and the API on the single JSON field / SQL column level. They can be created with `fields.NewField("field_name")`. The API is pretty straightforward, please consult the [godoc](https://pkg.go.dev/github.com/glothriel/grf/pkg/fields).

TLDR; you can:

* Set the field as read-only, write-only or read-write
* Set the InternalValue function, that will be used to transform the data from the API to format that can be stored in the database
* Set the Representation function, that will be used to transform the data from the database to the API response