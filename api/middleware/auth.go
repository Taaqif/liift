// Package middleware provides authentication middleware for Echo framework.
package middleware

import (
	"net/http"

	"liift/api/types"
	"liift/internal/database"
	"liift/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// RequireAuth returns a middleware that validates the JWT Bearer token and
// attaches the authenticated user to the request context.
func RequireAuth(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "authorization_header_missing",
				})
			}

			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.ErrUnauthorized
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "invalid_token",
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "invalid_token_claims",
				})
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "invalid_user_id_in_token",
				})
			}

			userID := uint(userIDFloat)

			var user models.User
			if err := database.DB.First(&user, userID).Error; err != nil {
				return c.JSON(http.StatusUnauthorized, types.ErrorResponse{
					Error: "user_not_found",
				})
			}

			c.Set("user", &user)
			c.Set("user_id", userID)
			c.Set("role", user.Role)

			return next(c)
		}
	}
}

// RequireAdmin returns a middleware that enforces admin role after RequireAuth.
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if role != "admin" {
				return c.JSON(http.StatusForbidden, types.ErrorResponse{
					Error: "admin_required",
				})
			}
			return next(c)
		}
	}
}

// GetUser retrieves the authenticated user from the Echo context.
func GetUser(c echo.Context) *models.User {
	user, _ := c.Get("user").(*models.User)
	return user
}

// GetUserID retrieves the authenticated user's ID from the Echo context.
func GetUserID(c echo.Context) uint {
	userID, _ := c.Get("user_id").(uint)
	return userID
}

// GetRole retrieves the authenticated user's role from the Echo context.
func GetRole(c echo.Context) string {
	role, _ := c.Get("role").(string)
	return role
}
