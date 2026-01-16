// Package api provides HTTP handlers and routes for the API endpoints.
package api

import (
	"liift/api/handlers"
	"liift/api/middleware"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterHandlers(e *echo.Echo, db *gorm.DB) {
	apiGroup := e.Group("/api")

	handlers.RegisterSystemRoutes(apiGroup)

	authHandler := handlers.NewAuthHandler(db)
	handlers.RegisterAuthRoutes(apiGroup, authHandler)

	protected := apiGroup.Group("", middleware.RequireAuth)

	// Create handlers with dependencies
	equipmentHandler := handlers.NewEquipmentHandler(db)
	handlers.RegisterEquipmentRoutes(protected, equipmentHandler)

	muscleGroupHandler := handlers.NewMuscleGroupHandler(db)
	handlers.RegisterMuscleGroupRoutes(protected, muscleGroupHandler)

	exerciseRepo := repository.NewExerciseRepository(db)
	exerciseHandler := handlers.NewExerciseHandler(exerciseRepo)
	handlers.RegisterExerciseRoutes(protected, exerciseHandler)
}
