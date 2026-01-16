package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// MuscleGroupHandler handles muscle group-related HTTP requests
type MuscleGroupHandler struct {
	db *gorm.DB
}

// NewMuscleGroupHandler creates a new MuscleGroupHandler with the given database connection
func NewMuscleGroupHandler(db *gorm.DB) *MuscleGroupHandler {
	return &MuscleGroupHandler{db: db}
}

func (h *MuscleGroupHandler) GetMuscleGroups(c echo.Context) error {
	var muscleGroups []models.MuscleGroup
	if err := h.db.Find(&muscleGroups).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to fetch muscle groups",
		})
	}

	names := make([]string, len(muscleGroups))
	for i, mg := range muscleGroups {
		names[i] = mg.Name
	}

	return c.JSON(http.StatusOK, names)
}

// RegisterMuscleGroupRoutes registers muscle group routes with the given handler
func RegisterMuscleGroupRoutes(api *echo.Group, handler *MuscleGroupHandler) {
	api.GET("/muscle-groups", handler.GetMuscleGroups)
}
