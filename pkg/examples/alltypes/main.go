package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BoolModel struct {
	models.BaseModel
	Value bool `json:"value" gorm:"column:value"`
}

type StringModel struct {
	models.BaseModel
	Value string `json:"value" gorm:"column:value"`
}

type StringPtrModel struct {
	models.BaseModel
	Value *string `json:"value" gorm:"column:value"`
}

type NullStringModel struct {
	models.BaseModel
	Value sql.NullString `json:"value" gorm:"column:value"`
}

type IntModel struct {
	models.BaseModel
	Value int `json:"value" gorm:"column:value"`
}
type UintModel struct {
	models.BaseModel
	Value uint `json:"value" gorm:"column:value"`
}

type FloatModel struct {
	models.BaseModel
	Value float64 `json:"value" gorm:"column:value"`
}

type DateTimeModel struct {
	models.BaseModel
	Value time.Time `json:"value" gorm:"column:value"`
}

type DurationModel struct {
	models.BaseModel
	Value time.Duration `json:"value" gorm:"column:value"`
}

type StringSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[string] `json:"value" gorm:"column:value;type:text"`
}

type IntSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[int] `json:"value" gorm:"column:value;type:text"`
}

type FloatSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[float64] `json:"value" gorm:"column:value;type:text"`
}

type MapSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[map[string]any] `json:"value" gorm:"column:value;type:text"`
}

type BoolSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[bool] `json:"value" gorm:"column:value;type:text"`
}

type AnySliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[any] `json:"value" gorm:"column:value;type:text"`
}

type TwoDStringSliceModel struct {
	models.BaseModel
	Value fields.SliceModelField[fields.SliceModelField[string]] `json:"value" gorm:"column:value;type:text"`
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "alltypes.db", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()
	gormDB, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	registerModel[BoolModel](router, gormDB, "/bool_field", "created_at")
	registerModel[StringModel](router, gormDB, "/string_field", "created_at")
	// registerModel[StringPtrModel](router, gormDB, "/string_pointer_field", "created_at")
	registerModel[IntModel](router, gormDB, "/int_field", "created_at")
	registerModel[UintModel](router, gormDB, "/uint_field", "created_at")
	registerModel[FloatModel](router, gormDB, "/float_field", "created_at")
	registerModel[StringSliceModel](router, gormDB, "/string_slice_field", "created_at")
	registerModel[FloatSliceModel](router, gormDB, "/float_slice_field", "created_at")
	registerModel[MapSliceModel](router, gormDB, "/map_slice_field", "created_at")
	registerModel[BoolSliceModel](router, gormDB, "/bool_slice_field", "created_at")

	// registerModel[DurationModel](router, gormDB, "/duration_field", "created_at")
	registerModel[DateTimeModel](router, gormDB, "/datetime_field", "created_at")
	registerModel[IntSliceModel](router, gormDB, "/int_slice_field", "created_at")
	// registerModel[NullStringModel](router, gormDB, "/null_string_field", "created_at")

	registerModel[AnySliceModel](router, gormDB, "/any_slice_field", "created_at")

	registerModel[TwoDStringSliceModel](router, gormDB, "/two_d_string_slice_field", "created_at")

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}

func registerModel[Model any](
	router *gin.Engine,
	gormDB *gorm.DB,
	prefix string,
	orderBy string,
) {
	serializer := serializers.NewModelSerializer[Model]()

	views.NewListCreateModelView[Model](prefix, queries.GORM[Model](gormDB).WithOrderBy(fmt.Sprintf("%s ASC", orderBy))).WithSerializer(
		serializer,
	).Register(router)

	views.NewRetrieveUpdateDestroyModelView[Model](prefix+"/:id", queries.GORM[Model](gormDB)).WithSerializer(
		serializer,
	).Register(router)

	var entity Model
	if migrateErr := gormDB.AutoMigrate(&entity); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
}
