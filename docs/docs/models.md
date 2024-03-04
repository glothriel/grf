# Models

In GRF, models are employed to specify the schema of the data that you want to read or write. For performance reasons, the data is internally handled as maps and only converted to model structs when required (e.g., in the GORM query driver during create and update operations). The `models.InternalValue` type serves as a map representation of the model across the framework.


## Introduction

### Requirements

:::info
To use GRF, you need to have a single primary key (either an `u?int[0-9]{2}` or a string) that is annotated with the `id` JSON tag. It's not possible to use composite primary keys, and you also can't use GRF without a primary key. 
:::

GRF only mandates the struct to have an exported field with an `id` JSON tag. However, depending on the Query Driver implementation, additional requirements may be necessary. For instance, if you intend to store your models in a SQL database (e.g., with the default GORM Query Driver), you must ensure that all fields implement the sql.Scanner and driver.Valuer interfaces.

```go
// this will work
type ValidModel struct {
    ID string `json:"id"`
}

// this will fail
type InvalidModel{
    Foo string
}
```

The `id` field requirement is non-negotiable, as it is used to uniquely identify the model in the storage. The `id` field can be either a numeric (assumed auto incremented ID field) or a string (assumed UUID, but any will be fine, as long as it's unique) type.


### Conversion between struct and `models.InternalValue`

GRF exposes two functions in `models` package, that allow conversion between struct and `models.InternalValue`:

* `models.AsInternalValue` - converts a struct to `models.InternalValue`
* `models.AsModel` - converts `models.InternalValue` to a struct

The functions use reflection to convert between the types, so they are not the fastest. However, they are very convenient, and you can always implement your own conversion functions, if the speed is an issue.


## Model fields


:::danger

Those types are not supported yet, but will be in the future:

* time.Time
* time.Duration
* `slice<int>`
* pointer fields, for example `*string`
* JSON fields, as non-string, dedicated column types (eg. Postgres)

:::

All the model fields must be tagged with a `json` tag, which is used to map the field name in `models.InternalValue`, and thus to (at least default) request and response JSON payloads. Saying that, GRF must know how to:

* Read the field from the storage (in case of GORM, the field should implement the sql.Scanner interface)
* Write the field to the storage (in case of GORM, the field should implement the driver.Valuer interface)
* Read the field from the request JSON payload (GRF uses some tricks to do that, read more in the [Serializers](./serializers) section)
* Write the field to the response JSON payload, read more in the [Serializers](./serializers) section, like above.


### Slice field

`models.SliceField` can be used to store slices, that are encoded to JSON string for storage (implement sql.Scanner and driver.Valuer interfaces). In request and response JSON payloads, the slice is represented as a JSON array. The types of the slice need to be golang built-in types. The field provides validation of all the elements in the slice.

