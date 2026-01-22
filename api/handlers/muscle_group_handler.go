package handlers

import (
	"net/http"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MuscleGroupHandler struct {
	db *gorm.DB
}

func NewMuscleGroupHandler(db *gorm.DB) *MuscleGroupHandler {
	return &MuscleGroupHandler{db: db}
}

type GetMuscleGroupsResponse struct {
	Name string `json:"name"`
}

func (h *MuscleGroupHandler) GetMuscleGroups(c echo.Context) error {
	var muscleGroups []models.MuscleGroup
	if err := h.db.Find(&muscleGroups).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "muscle_groups_fetch_failed",
		})
	}

	muscleGroupsResponse := utils.Map(muscleGroups, func(u models.MuscleGroup) GetMuscleGroupsResponse {
		return GetMuscleGroupsResponse{
			Name: u.Name,
		}
	})

	return c.JSON(http.StatusOK, muscleGroupsResponse)
}

func RegisterMuscleGroupRoutes(api *echo.Group, handler *MuscleGroupHandler) {
	api.GET("/muscle-groups", handler.GetMuscleGroups)
}
