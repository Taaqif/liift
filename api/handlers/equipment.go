package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// EquipmentHandler handles equipment-related HTTP requests
type EquipmentHandler struct {
	db *gorm.DB
}

// NewEquipmentHandler creates a new EquipmentHandler with the given database connection
func NewEquipmentHandler(db *gorm.DB) *EquipmentHandler {
	return &EquipmentHandler{db: db}
}

func (h *EquipmentHandler) GetEquipment(c echo.Context) error {
	var equipment []models.Equipment
	if err := h.db.Find(&equipment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to fetch equipment",
		})
	}

	names := make([]string, len(equipment))
	for i, e := range equipment {
		names[i] = e.Name
	}

	return c.JSON(http.StatusOK, names)
}

// RegisterEquipmentRoutes registers equipment routes with the given handler
func RegisterEquipmentRoutes(api *echo.Group, handler *EquipmentHandler) {
	api.GET("/equipment", handler.GetEquipment)
}
