package handlers

import (
	"net/http"
	"time"

	"liift/api/types"
	"liift/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"omitempty,email"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret []byte
}

func NewAuthHandler(db *gorm.DB, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "invalid_credentials",
		})
	}

	if !user.CheckPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "invalid_credentials",
		})
	}

	token, err := h.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "token_generation_failed",
		})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	var existingUser models.User
	query := h.db.Where("username = ?", req.Username)
	if req.Email != "" {
		query = query.Or("email = ?", req.Email)
	}
	if err := query.First(&existingUser).Error; err == nil {
		return c.JSON(http.StatusConflict, types.ErrorResponse{
			Error: "user_already_exists",
		})
	}

	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}
	// First registered user becomes admin
	var userCount int64
	h.db.Model(&models.User{}).Count(&userCount)
	role := "user"
	if userCount == 0 {
		role = "admin"
	}

	user := models.User{
		Username: req.Username,
		Email:    emailPtr,
		Role:     role,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to hash password",
		})
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "user_creation_failed",
		})
	}

	token, err := h.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "token_generation_failed",
		})
	}

	return c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) generateToken(userID uint, username string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func RegisterAuthRoutes(api *echo.Group, handler *AuthHandler) {
	api.POST("/auth/login", handler.Login)
	api.POST("/auth/register", handler.Register)
}
