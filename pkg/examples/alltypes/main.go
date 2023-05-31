package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/gin-rest-framework/pkg/fields"
	"github.com/glothriel/gin-rest-framework/pkg/models"
	"github.com/glothriel/gin-rest-framework/pkg/serializers"
	"github.com/glothriel/gin-rest-framework/pkg/views"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BoolField struct {
	models.BaseModel
	Value bool `json:"value" gorm:"column:value"`
}

type StringField struct {
	models.BaseModel
	Value string `json:"value" gorm:"column:value"`
}

type StringPointerField struct {
	models.BaseModel
	Value *string `json:"value" gorm:"column:value"`
}

type NullStringField struct {
	models.BaseModel
	Value sql.NullString `json:"value" gorm:"column:value"`
}

type IntField struct {
	models.BaseModel
	Value int `json:"value" gorm:"column:value"`
}
type UintField struct {
	models.BaseModel
	Value uint `json:"value" gorm:"column:value"`
}

type FloatField struct {
	models.BaseModel
	Value float64 `json:"value" gorm:"column:value"`
}

type DateTimeField struct {
	models.BaseModel
	Value time.Time `json:"value" gorm:"column:value"`
}

type DurationField struct {
	models.BaseModel
	Value time.Duration `json:"value" gorm:"column:value"`
}

type StringSliceField struct {
	models.BaseModel
	Value fields.SliceModelField[string, StringSliceField] `json:"value" gorm:"column:value;type:text"`
}

type IntSliceField struct {
	models.BaseModel
	Value fields.SliceModelField[int, IntSliceField] `json:"value" gorm:"column:value;type:text"`
}

type FloatSliceField struct {
	models.BaseModel
	Value fields.SliceModelField[float64, FloatSliceField] `json:"value" gorm:"column:value;type:text"`
}

type MapSliceField struct {
	models.BaseModel
	Value fields.SliceModelField[map[string]any, MapSliceField] `json:"value" gorm:"column:value;type:text"`
}

type BoolSliceField struct {
	models.BaseModel
	Value fields.SliceModelField[bool, BoolSliceField] `json:"value" gorm:"column:value;type:text"`
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "alltypes.db", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()
	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	registerModel[BoolField](router, db, "/bool_field", "created_at")
	registerModel[StringField](router, db, "/string_field", "created_at")
	registerModel[StringPointerField](router, db, "/string_pointer_field", "created_at")
	registerModel[IntField](router, db, "/int_field", "created_at")
	registerModel[UintField](router, db, "/uint_field", "created_at")
	registerModel[FloatField](router, db, "/float_field", "created_at")
	registerModel[StringSliceField](router, db, "/string_slice_field", "created_at")
	registerModel[FloatSliceField](router, db, "/float_slice_field", "created_at")
	registerModel[MapSliceField](router, db, "/map_slice_field", "created_at")
	registerModel[BoolSliceField](router, db, "/bool_slice_field", "created_at")

	registerModel[DurationField](router, db, "/duration_field", "created_at")
	registerModel[DateTimeField](router, db, "/datetime_field", "created_at")
	registerModel[IntSliceField](router, db, "/int_slice_field", "created_at")
	registerModel[NullStringField](router, db, "/null_string_field", "created_at")

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}

func registerModel[Model any](
	router *gin.Engine,
	db *gorm.DB,
	prefix string,
	orderBy string,
) {
	serializer := serializers.NewValidatingSerializer[Model](
		serializers.NewModelSerializer[Model](nil)).WithValidator(
		&serializers.GoPlaygroundValidator[Model]{},
	)

	views.NewListCreateModelView[Model](prefix, db).WithSerializer(
		serializer,
	).WithOrderBy(fmt.Sprintf("%s ASC", orderBy)).Register(router)

	views.NewRetrieveUpdateDeleteModelView[Model](prefix+"/:id", db).WithSerializer(
		serializer,
	).Register(router)

	var entity Model
	if migrateErr := db.AutoMigrate(&entity); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}
}
