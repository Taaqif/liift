package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ExerciseFeatureHandler struct {
	db *gorm.DB
}

func NewExerciseFeatureHandler(db *gorm.DB) *ExerciseFeatureHandler {
	return &ExerciseFeatureHandler{db: db}
}

type GetExerciseFeatureResponse struct {
	Name string `json:"name"`
}

func (h *ExerciseFeatureHandler) GetExerciseFeatures(c echo.Context) error {
	var exerciseFeatures []models.ExerciseFeature
	if err := h.db.Find(&exerciseFeatures).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_feature_fetch_failed",
		})
	}

	exerciseFeatureResponse := utils.Map(exerciseFeatures, func(u models.ExerciseFeature) GetExerciseFeatureResponse {
		return GetExerciseFeatureResponse{
			Name: u.Name,
		}
	})

	return c.JSON(http.StatusOK, exerciseFeatureResponse)
}

func RegisterExerciseFeatureRoutes(api *echo.Group, handler *ExerciseFeatureHandler) {
	api.GET("/exercise-features", handler.GetExerciseFeatures)
}
