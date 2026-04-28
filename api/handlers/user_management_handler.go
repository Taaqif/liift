package handlers

import (
	"net/http"
	"strconv"

	"liift/api/middleware"
	"liift/api/types"
	"liift/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UserManagementHandler struct {
	db *gorm.DB
}

func NewUserManagementHandler(db *gorm.DB) *UserManagementHandler {
	return &UserManagementHandler{db: db}
}

type UserListItem struct {
	ID       uint    `json:"id"`
	Username string  `json:"username"`
	Email    *string `json:"email"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
}

func (h *UserManagementHandler) ListUsers(c echo.Context) error {
	var users []models.User
	if err := h.db.Order("id asc").Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_users"})
	}
	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Name:     u.Name,
			Role:     u.Role,
		}
	}
	return c.JSON(http.StatusOK, items)
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

func (h *UserManagementHandler) UpdateUserRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_user_id"})
	}

	// Prevent self-demotion
	callerID := middleware.GetUserID(c)
	if uint(id) == callerID {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "cannot_change_own_role"})
	}

	var req UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}
	if req.Role != "admin" && req.Role != "user" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_role"})
	}

	var user models.User
	if err := h.db.First(&user, uint(id)).Error; err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "user_not_found"})
	}

	if err := h.db.Model(&user).Update("role", req.Role).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_update_role"})
	}

	return c.JSON(http.StatusOK, UserListItem{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Role:     req.Role,
	})
}

func (h *UserManagementHandler) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_user_id"})
	}

	// Prevent self-deletion
	callerID := middleware.GetUserID(c)
	if uint(id) == callerID {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "cannot_delete_own_account"})
	}

	if err := h.db.Delete(&models.User{}, uint(id)).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_delete_user"})
	}

	return c.NoContent(http.StatusNoContent)
}

func RegisterUserManagementRoutes(api *echo.Group, handler *UserManagementHandler) {
	admin := api.Group("", middleware.RequireAdmin())
	admin.GET("/users", handler.ListUsers)
	admin.PUT("/users/:id/role", handler.UpdateUserRole)
	admin.DELETE("/users/:id", handler.DeleteUser)
}
