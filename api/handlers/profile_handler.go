package handlers

import (
	"net/http"

	"liift/api/middleware"
	"liift/api/types"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UpdateProfileRequest struct {
	Name               *string  `json:"name"`
	DateOfBirth        *string  `json:"date_of_birth"`
	Gender             *string  `json:"gender"`
	WeightKg           *float64 `json:"weight_kg"`
	HeightCm           *float64 `json:"height_cm"`
	OnboardingComplete *bool    `json:"onboarding_complete"`
}

type ProfileResponse struct {
	ID                 uint     `json:"id"`
	Username           string   `json:"username"`
	Email              *string  `json:"email"`
	Name               string   `json:"name"`
	DateOfBirth        string   `json:"date_of_birth"`
	Gender             string   `json:"gender"`
	WeightKg           *float64 `json:"weight_kg"`
	HeightCm           *float64 `json:"height_cm"`
	OnboardingComplete bool     `json:"onboarding_complete"`
}

type ProfileHandler struct {
	db *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func (h *ProfileHandler) GetMe(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "unauthorized"})
	}
	return c.JSON(http.StatusOK, ProfileResponse{
		ID:                 user.ID,
		Username:           user.Username,
		Email:              user.Email,
		Name:               user.Name,
		DateOfBirth:        user.DateOfBirth,
		Gender:             user.Gender,
		WeightKg:           user.WeightKg,
		HeightCm:           user.HeightCm,
		OnboardingComplete: user.OnboardingComplete,
	})
}

func (h *ProfileHandler) UpdateMe(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "unauthorized"})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = *req.DateOfBirth
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.WeightKg != nil {
		user.WeightKg = req.WeightKg
	}
	if req.HeightCm != nil {
		user.HeightCm = req.HeightCm
	}
	if req.OnboardingComplete != nil {
		user.OnboardingComplete = *req.OnboardingComplete
	}

	if err := h.db.Save(user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "update_failed"})
	}

	return c.JSON(http.StatusOK, ProfileResponse{
		ID:                 user.ID,
		Username:           user.Username,
		Email:              user.Email,
		Name:               user.Name,
		DateOfBirth:        user.DateOfBirth,
		Gender:             user.Gender,
		WeightKg:           user.WeightKg,
		HeightCm:           user.HeightCm,
		OnboardingComplete: user.OnboardingComplete,
	})
}

func RegisterProfileRoutes(api *echo.Group, handler *ProfileHandler) {
	api.GET("/users/me", handler.GetMe)
	api.PUT("/users/me", handler.UpdateMe)
}
