// Package api provides HTTP handlers and routes for the API endpoints.
package api

import (
	"liift/api/handlers"
	"liift/api/middleware"
	"liift/internal/repository"
	"liift/internal/utils"

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

	exerciseFeatureHandler := handlers.NewExerciseFeatureHandler(db)
	handlers.RegisterExerciseFeatureRoutes(protected, exerciseFeatureHandler)

	muscleGroupHandler := handlers.NewMuscleGroupHandler(db)
	handlers.RegisterMuscleGroupRoutes(protected, muscleGroupHandler)

	imageStoragePath := utils.GetEnv("IMAGE_STORAGE_PATH", "./storage/images")
	imageRepo := repository.NewImageRepository(db)
	imageHandler := handlers.NewImageHandler(imageRepo, imageStoragePath)
	handlers.RegisterImageRoutes(protected, imageHandler)

	exerciseRepo := repository.NewExerciseRepository(db)
	exerciseHandler := handlers.NewExerciseHandler(exerciseRepo, imageRepo, imageStoragePath)
	handlers.RegisterExerciseRoutes(protected, exerciseHandler)
}
