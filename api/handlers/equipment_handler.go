package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type EquipmentHandler struct {
	db *gorm.DB
}

func NewEquipmentHandler(db *gorm.DB) *EquipmentHandler {
	return &EquipmentHandler{db: db}
}

type GetEquipmentResponse struct {
	Name string `json:"name"`
}

func (h *EquipmentHandler) GetEquipment(c echo.Context) error {
	var equipment []models.Equipment
	if err := h.db.Find(&equipment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "equipment_fetch_failed",
		})
	}

	equipmentResponse := utils.Map(equipment, func(u models.Equipment) GetEquipmentResponse {
		return GetEquipmentResponse{
			Name: u.Name,
		}
	})

	return c.JSON(http.StatusOK, equipmentResponse)
}

func RegisterEquipmentRoutes(api *echo.Group, handler *EquipmentHandler) {
	api.GET("/equipment", handler.GetEquipment)
}
