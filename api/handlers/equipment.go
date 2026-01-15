package handlers

import (
	"net/http"

	"liift/internal/database"
	"liift/internal/models"

	"github.com/labstack/echo/v4"
)

func GetEquipment(c echo.Context) error {
	var equipment []models.Equipment
	if err := database.DB.Find(&equipment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch equipment",
		})
	}

	names := make([]string, len(equipment))
	for i, e := range equipment {
		names[i] = e.Name
	}

	return c.JSON(http.StatusOK, names)
}

func RegisterEquipmentRoutes(api *echo.Group) {
	api.GET("/equipment", GetEquipment)
}
