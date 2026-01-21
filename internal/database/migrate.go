package database

import (
	"log"

	"liift/internal/models"

	"gorm.io/gorm"
)

// Migrate runs database migrations for all models
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Equipment{},
		&models.MuscleGroup{},
		&models.Image{},
		&models.Exercise{},
		&models.User{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateModels runs migrations for specific models
func MigrateModels(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}

	log.Printf("Running migrations for %d model(s)...", len(models))
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}
	log.Println("Migrations completed successfully")
	return nil
}
