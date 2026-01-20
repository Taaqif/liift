// Package middleware provides authentication middleware for Echo framework.
package middleware

import (
	"net/http"

	"liift/api/types"
	"liift/internal/database"
	"liift/internal/models"
	"liift/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecret []byte

func init() {
	secret := utils.GetEnv("JWT_SECRET", "")
	jwtSecret = []byte(secret)
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
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
			return jwtSecret, nil
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

		return next(c)
	}
}

func GetUser(c echo.Context) *models.User {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetUserID(c echo.Context) uint {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return 0
	}
	return userID
}
