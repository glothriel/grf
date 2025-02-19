# Query drivers

Query drivers are the storage layer of GRF. They are responsible for storing and retrieving data from the database (or other source), and they are also responsible for translating the query parameters into the database query (filtering, sorting, pagination). QueryDriver can also hook into Gin request lifecycle by providing middleware - this is useful for example to initialize query object in request context when the request is received.

For most cases you should be fine with using GORM query driver, but consider reviewing the [Writing own query driver](#writing-own-query-driver) section if it's not sufficient for you.


## Usage

### GORM `queries.GORM(*gorm.DB)`

GORM query driver is the only production-ready driver included in GRF. It uses GORM models to store data in any database supported by GORM. It supports:

* filtering (`driver.WithFilter`)
* sorting (`driver.WithOrderBy`) 
* pagination (`driver.WithPagination`)

Here's an example of using GORM query driver (taken from `pkg/exammples/products` package):

```go
import(
    ...
	"github.com/glothriel/grf/pkg/queries/gormq"
    ...
)

...

type Product struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
    Name  string `json:"name"`
    Price float64 `json:"price"`
}

...
// Setup *gorm.DB
gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
if err != nil {
    panic("failed to connect database")
}
if migrateErr := gormDB.AutoMigrate(&Product{}); migrateErr != nil {
    panic(fmt.Sprintf("Error migrating database: %s", migrateErr))
}

// Declare query driver
queries.GORM[Product](gormDB).WithFilter(
    func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
        if ctx.Query("name") != "" {
            return db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("name")))
        }
        return db
    },
).WithOrderBy("name ASC").WithPagination(&gormq.LimitOffsetPagination{})
...
```

Driver configured in such way:

* Allows filtering the list of products by name (using `name` query parameter)
* Sorts the list of products by name in ascending order
* Uses limit/offset pagination provided by gorm query driver package

#### Transactions

All the default REST actions are performed in a single query, thus a transaction is not strictly needed. If however you'd like your action to have some side-effects (for example saving an entry in an audit log), you can use GORM query driver's transaction support.

```go
queryDriver.CRUD().WithCreate(
    gormq.CreateTx(
        gormq.BeforeCreate(
            func(ctx *gin.Context, iv models.InternalValue, tx *gorm.DB) (models.InternalValue, error) {
                // do whatever you want with tx before creating
                return iv, nil
            },
        ),
    )(queryDriver.CRUD().Create),
)
```

The API is a little bit complex (with functions returning functions creating functions 🤣), so it may be changed at some point, but for now it does the job.

#### Relationships

GORM query driver supports basic relationships between models. See more in [model relations section](./models#model-relations).

### InMemory `queries.InMemory()`

InMemory query driver is a simple implementation of QueryDriver interface, that stores all the data in memory. It's useful for testing and prototyping, but it definetly should not be used in production. It doesn't support any filtering, sorting or pagination.

## Writing own query driver

You may consider writing your own query driver if:

* You'd like to use GRF with other ORM or query builder
* You'd like to use GRF with other database than those supported by GORM (non-relational included - in theory you could use GRF with any database, but you'd have to implement your own query driver)
* If you'd like to create a wrapper around other API (for example REST API, XML, etc) and use GRF to query it

Implementation of own query driver is straightforward - you just have to implement the `queries.Driver` interface. To kick-start your implementation, you can use the `queries.InMemory` driver as a reference. Important things to keep in mind while implementing:

* If you need something to be done during request lifecycle (for example before or after request) use Gin middlewares
* If you need to pass something to/from your application code and query driver, use Gin context. The method that works the best is to include `CtxGetSomething(*gin.Context)` - like methods alongside your implementation, so you can quickly use whatever your middleware has set up for you in other parts of your code. You can see an example of this in `queries.GORM` driver, where subsequent calls to `CtxQuery` return the same query builder, that is modified by filter, pagination or sorting mechanisms.