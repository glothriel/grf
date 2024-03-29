package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
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
	Value models.SliceField[string] `json:"value" gorm:"column:value;type:text"`
}

type IntSliceModel struct {
	models.BaseModel
	Value models.SliceField[int] `json:"value" gorm:"column:value;type:text"`
}

type FloatSliceModel struct {
	models.BaseModel
	Value models.SliceField[float64] `json:"value" gorm:"column:value;type:text"`
}

type MapSliceModel struct {
	models.BaseModel
	Value models.SliceField[map[string]any] `json:"value" gorm:"column:value;type:text"`
}

type BoolSliceModel struct {
	models.BaseModel
	Value models.SliceField[bool] `json:"value" gorm:"column:value;type:text"`
}

type AnySliceModel struct {
	models.BaseModel
	Value models.SliceField[any] `json:"value" gorm:"column:value;type:text"`
}

type TwoDStringSliceModel struct {
	models.BaseModel
	Value models.SliceField[models.SliceField[string]] `json:"value" gorm:"column:value;type:text"`
}

type NullBoolModel struct {
	models.BaseModel
	Value sql.NullBool `json:"value" gorm:"column:value"`
}

type NullInt16Model struct {
	models.BaseModel
	Value sql.NullInt16 `json:"value" gorm:"column:value"`
}

type NullInt32Model struct {
	models.BaseModel
	Value sql.NullInt32 `json:"value" gorm:"column:value"`
}

type NullInt64Model struct {
	models.BaseModel
	Value sql.NullInt64 `json:"value" gorm:"column:value"`
}
type NullFloat64Model struct {
	models.BaseModel
	Value sql.NullFloat64 `json:"value" gorm:"column:value"`
}

type NullStringModel struct {
	models.BaseModel
	Value sql.NullString `json:"value" gorm:"column:value"`
}

type NullTimeModel struct {
	models.BaseModel
	Value sql.NullTime `json:"value" gorm:"column:value"`
}

type NullByteModel struct {
	models.BaseModel
	Value sql.NullByte `json:"value" gorm:"column:value"`
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

	registerModel[IntModel](router, gormDB, "/int_field", "created_at")
	registerModel[UintModel](router, gormDB, "/uint_field", "created_at")
	registerModel[FloatModel](router, gormDB, "/float_field", "created_at")
	registerModel[StringSliceModel](router, gormDB, "/string_slice_field", "created_at")
	registerModel[FloatSliceModel](router, gormDB, "/float_slice_field", "created_at")
	registerModel[MapSliceModel](router, gormDB, "/map_slice_field", "created_at")
	registerModel[BoolSliceModel](router, gormDB, "/bool_slice_field", "created_at")

	registerModel[DateTimeModel](router, gormDB, "/datetime_field", "created_at")
	registerModel[IntSliceModel](router, gormDB, "/int_slice_field", "created_at")

	registerModel[AnySliceModel](router, gormDB, "/any_slice_field", "created_at")

	registerModel[TwoDStringSliceModel](router, gormDB, "/two_d_string_slice_field", "created_at")
	registerModel[NullBoolModel](router, gormDB, "/null_bool_field", "created_at")
	registerModel[NullStringModel](router, gormDB, "/null_string_field", "created_at")
	registerModel[NullInt16Model](router, gormDB, "/null_int16_field", "created_at")
	registerModel[NullInt32Model](router, gormDB, "/null_int32_field", "created_at")
	registerModel[NullInt64Model](router, gormDB, "/null_int64_field", "created_at")
	registerModel[NullFloat64Model](router, gormDB, "/null_float64_field", "created_at")
	registerModel[NullByteModel](router, gormDB, "/null_byte_field", "created_at")

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}

func registerModel[Model any](
	router *gin.Engine,
	gormDB *gorm.DB,
	prefix string,
	orderBy string,
) {

	views.NewModelViewSet[Model](prefix, queries.GORM[Model](gormDB).WithOrderBy(fmt.Sprintf("%s ASC", orderBy))).Register(router)

	var entity Model
	if migrateErr := gormDB.AutoMigrate(&entity); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
}
