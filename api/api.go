// Package api provides HTTP handlers and routes for the API endpoints.
package api

import (
	"liift/api/handlers"
	"liift/api/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo) {
	apiGroup := e.Group("/api")

	handlers.RegisterSystemRoutes(apiGroup)
	handlers.RegisterAuthRoutes(apiGroup)

	protected := apiGroup.Group("", middleware.RequireAuth)
	handlers.RegisterEquipmentRoutes(protected)
	handlers.RegisterMuscleGroupRoutes(protected)
	handlers.RegisterExerciseRoutes(protected)
}
