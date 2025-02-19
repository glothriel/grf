package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries"
	"github.com/glothriel/grf/pkg/serializers"
	"github.com/glothriel/grf/pkg/views"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Profile struct {
	models.BaseModel
	Name   string  `json:"name" gorm:"size:191;column:name"`
	Photos []Photo `json:"photos" gorm:"foreignKey:profile_id" grf:"relation"`
}

type Photo struct {
	models.BaseModel
	ProfileID uuid.UUID `json:"profile_id" gorm:"size:191;column:profile_id"`
}

func SeedData(db *gorm.DB) ([]string, error) {
	// Create two profiles
	profiles := []Profile{
		{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Name:      "Kajtek",
		},
		{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Name:      "Roksana",
		},
	}
	uuids := make([]string, 0, len(profiles))
	if err := db.Create(&profiles).Error; err != nil {
		return uuids, fmt.Errorf("failed to create profiles: %w", err)
	}
	uuids = []string{profiles[0].ID.String(), profiles[1].ID.String()}

	photos := []Photo{
		{
			ProfileID: profiles[0].ID,
		},
		{
			ProfileID: profiles[0].ID,
		},
		{
			ProfileID: profiles[0].ID,
		},
		{
			ProfileID: profiles[1].ID,
		},
		{
			ProfileID: profiles[1].ID,
		},
		{
			ProfileID: profiles[1].ID,
		},
	}

	// Insert photos
	if err := db.Create(&photos).Error; err != nil {
		return uuids, fmt.Errorf("failed to create photos: %w", err)
	}

	return uuids, nil
}

func TestRelations(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	gormDB, gormOpenErr := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, gormOpenErr)

	views.NewModelViewSet[Profile](
		"/profiles",
		queries.GORM[Profile](
			gormDB,
		).WithPreload(
			"photos",
		).WithOrderBy(
			"`profiles`.`created_at` ASC",
		),
	).WithSerializer(
		serializers.NewModelSerializer[Profile]().WithNewField(
			serializers.NewSerializerField[Photo](
				"photos",
				serializers.NewModelSerializer[Photo](),
			),
		),
	).Register(router)
	autoMigrateErr := gormDB.AutoMigrate(Profile{}, Photo{})
	profileIDs, seedErr := SeedData(gormDB)

	require.NoError(t, autoMigrateErr)
	require.NoError(t, seedErr)

	newRequestTestCase(t, "relations").Req(
		newRequest("GET", "/profiles", nil),
	).ExCode(
		http.StatusOK,
	).ExJson(
		[]any{
			map[string]any{
				"name": "Kajtek",
				"photos": []any{
					map[string]any{
						"profile_id": profileIDs[0],
					},
					map[string]any{
						"profile_id": profileIDs[0],
					},
					map[string]any{
						"profile_id": profileIDs[0],
					},
				},
			},
			map[string]any{
				"name": "Roksana",
				"photos": []any{
					map[string]any{
						"profile_id": profileIDs[1],
					},
					map[string]any{
						"profile_id": profileIDs[1],
					},
					map[string]any{
						"profile_id": profileIDs[1],
					},
				},
			},
		},
	).Run(router)
}
