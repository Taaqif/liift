// Package api provides HTTP handlers and routes for the API endpoints.
package api

import (
	"liift/api/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo) {
	apiGroup := e.Group("/api")

	handlers.RegisterSystemRoutes(apiGroup)
	handlers.RegisterEquipmentRoutes(apiGroup)
	handlers.RegisterMuscleGroupRoutes(apiGroup)
	handlers.RegisterExerciseRoutes(apiGroup)
}
