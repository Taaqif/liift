package api

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(api *echo.Group) {
	// Basic API endpoint
	api.GET("/message", MessageHandler)
}
