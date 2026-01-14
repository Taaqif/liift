// Package api provides HTTP handlers and routes for the API endpoints.
package api

import (
	"github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo) {
	// Setup the API Group
	apiGroup := e.Group("/api")

	// Register all API routes
	RegisterRoutes(apiGroup)
}
