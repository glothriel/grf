package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/queries/common"
	"github.com/glothriel/grf/pkg/queries/gormq"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Category struct {
	models.BaseModel

	Name string `json:"name" gorm:"size:191;column:name"`
}

type Product struct {
	models.BaseModel

	Name        string          `json:"name" gorm:"size:191;column:name"`
	Description string          `json:"description" gorm:"type:text;column:description"`
	Price       decimal.Decimal `json:"price" gorm:"type:decimal(19,4)"`

	CategoryID string   `json:"category_id" gorm:"size:191;column:category_id"`
	Category   Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null;"`
}

type Photo struct {
	models.BaseModel

	URL string `json:"url" gorm:"size:191;column:url"`

	ProductID string  `json:"product_id" gorm:"size:191;column:product_id"`
	Product   Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null;"`
}

type CustomerProfile struct {
	models.BaseModel

	Email     string `json:"email" gorm:"size:191;column:email;uniqueIndex"`
	FirstName string `json:"first_name" gorm:"size:191;column:first_name"`
	LastName  string `json:"last_name" gorm:"size:191;column:last_name"`
}

func main() {
	serverPort := flag.Int("port", 8080, "the port test server runs on")
	dbFile := flag.String("db", "file:/tmp/products.db?_foreign_keys=on", "the database file (sqlite) to use, will be created if not exists")
	flag.Parse()

	router := gin.Default()

	gormDB, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if migrateErr := gormDB.AutoMigrate(&Product{}, &Photo{}, &Category{}, &CustomerProfile{}); migrateErr != nil {
		logrus.Fatalf("Error migrating database: %s", migrateErr)
	}

	views.NewModelViewSet[Product](
		"/products",
		queries.GORM[Product](gormDB).WithFilter(
			func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				if ctx.Query("name") != "" {
					return db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", ctx.Query("name")))
				}
				return db
			},
		).WithOrderBy("name ASC"),
	).WithSerializer(
		serializers.NewValidatingSerializer[Product](
			serializers.NewModelSerializer[Product](),
			serializers.NewGoPlaygroundValidator[Product](
				map[string]any{
					"name":        "required",
					"description": "required",
					"category_id": "required",
				},
			),
		),
	).WithListSerializer(
		serializers.NewModelSerializer[Product]().
			WithModelFields([]string{"id", "name"}),
	).Register(router)

	views.NewViewSet[Category](
		"/categories",
		queries.GORM[Category](gormDB).WithOrderBy("name ASC"),
	).WithActions(
		views.ActionList, views.ActionCreate,
	).WithSerializer(
		serializers.NewValidatingSerializer[Category](
			serializers.NewModelSerializer[Category](),
			serializers.NewGoPlaygroundValidator[Category](
				map[string]any{
					"name": "required",
				},
			),
		),
	).WithListSerializer(
		serializers.NewModelSerializer[Category]().
			WithModelFields([]string{"id", "name"}),
	).Register(router)

	views.NewModelViewSet[Photo](
		"/products/:product_id/photos",
		queries.GORM[Photo](gormDB).WithFilter(
			func(ctx *gin.Context, db *gorm.DB) *gorm.DB {
				return db.Where("product_id = ?", ctx.Param("product_id"))
			},
		).WithOrderBy("created_at DESC"),
	).WithActions(
		views.ActionList,
	).WithSerializer(
		serializers.NewValidatingSerializer[Photo](
			serializers.NewModelSerializer[Photo]().WithField(
				"product_id",
				func(oldField *fields.Field[Photo]) {
					oldField.WriteOnly().WithInternalValueFunc(
						func(m map[string]any, s string, ctx *gin.Context) (any, error) {
							return ctx.Param("product_id"), nil
						},
					)
				},
			),
			serializers.NewGoPlaygroundValidator[Photo](
				map[string]any{
					"url": "required",
				},
			),
		),
	).Register(router)

	meQD := queries.GORM[CustomerProfile](gormDB)
	converter := gormq.FromDBConverter[CustomerProfile]()
	meQD.CRUD().WithRetrieve(
		func(ctx *gin.Context, id any) (models.InternalValue, error) {
			email := ctx.Request.Header.Get("X-User-Email")
			if email == "" {
				return nil, errors.New("X-User-Email header is required")
			}
			var empty CustomerProfile
			var rawEntity map[string]any
			retrieveErr := gormq.CtxQuery(ctx).Model(&empty).First(&rawEntity, "email = ?", email).Error
			if retrieveErr != nil {
				if retrieveErr == gorm.ErrRecordNotFound {
					return nil, common.ErrorNotFound
				}
				return nil, retrieveErr
			}
			return converter(rawEntity)
		},
	).WithCreate(
		func(ctx *gin.Context, m models.InternalValue) (models.InternalValue, error) {
			entity, asModelErr := models.AsModel[CustomerProfile](m)

			if asModelErr != nil {
				return nil, asModelErr
			}
			createErr := gormq.CtxQuery(ctx).Model(&entity).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "email"}},
				DoUpdates: clause.Assignments(m),
			}).Create(&entity).Error
			return models.AsInternalValue(entity), createErr
		},
	)

	views.NewViewSet[CustomerProfile](
		"/me",
		meQD,
	).WithExtraAction(
		views.NewExtraAction[CustomerProfile](
			"GET",
			"",
			views.RetrieveModelViewSetFunc[CustomerProfile],
		),
		serializers.NewModelSerializer[CustomerProfile]().WithModelFields([]string{"email", "last_name", "first_name"}),
		false,
	).WithExtraAction(
		views.NewExtraAction[CustomerProfile](
			"PUT",
			"",
			views.CreateModelViewSetFunc[CustomerProfile],
		),
		serializers.NewModelSerializer[CustomerProfile]().WithModelFields(
			[]string{"email", "last_name", "first_name"},
		).WithField(
			"email",
			func(oldField *fields.Field[CustomerProfile]) {
				oldField.WithInternalValueFunc(
					func(m map[string]any, s string, ctx *gin.Context) (any, error) {
						email := ctx.Request.Header.Get("X-User-Email")
						if email == "" {
							return nil, errors.New("X-User-Email header is required")
						}
						return email, nil
					},
				)
			},
		),
		false,
	).Register(router)

	logrus.Fatal(router.Run(fmt.Sprintf(":%d", *serverPort)))
}
