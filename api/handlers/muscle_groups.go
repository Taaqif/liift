package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/database"
	"liift/internal/models"

	"github.com/labstack/echo/v4"
)

func GetMuscleGroups(c echo.Context) error {
	var muscleGroups []models.MuscleGroup
	if err := database.DB.Find(&muscleGroups).Error; err != nil {
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

func RegisterMuscleGroupRoutes(api *echo.Group) {
	api.GET("/muscle-groups", GetMuscleGroups)
}
