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

func RegisterHandlers(e *echo.Echo, db *gorm.DB, jwtSecret []byte) {
	apiGroup := e.Group("/api")

	handlers.RegisterSystemRoutes(apiGroup)

	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	handlers.RegisterAuthRoutes(apiGroup, authHandler)

	protected := apiGroup.Group("", middleware.RequireAuth(jwtSecret))

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

	workoutRepo := repository.NewWorkoutRepository(db)
	workoutHandler := handlers.NewWorkoutHandler(workoutRepo)
	handlers.RegisterWorkoutRoutes(protected, workoutHandler)

	workoutPlanRepo := repository.NewWorkoutPlanRepository(db)
	workoutPlanHandler := handlers.NewWorkoutPlanHandler(workoutPlanRepo)
	handlers.RegisterWorkoutPlanRoutes(protected, workoutPlanHandler)

	workoutPlanProgressRepo := repository.NewWorkoutPlanProgressRepository(db)

	workoutSessionRepo := repository.NewWorkoutSessionRepository(db)
	workoutSessionHandler := handlers.NewWorkoutSessionHandler(workoutSessionRepo, workoutPlanProgressRepo)
	handlers.RegisterWorkoutSessionRoutes(protected, workoutSessionHandler)

	workoutPlanProgressHandler := handlers.NewWorkoutPlanProgressHandler(workoutPlanProgressRepo, workoutPlanRepo, workoutSessionRepo)
	handlers.RegisterWorkoutPlanProgressRoutes(protected, workoutPlanProgressHandler)
}
