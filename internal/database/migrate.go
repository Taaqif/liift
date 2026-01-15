// Package database provides database migration functionality.
package database

import (
	"log"

	"gorm.io/gorm"
)

// Migrate runs database migrations for all models
// Add your models here as you create them
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Example: Add your models here
	// Uncomment and modify as needed:
	// err := db.AutoMigrate(
	// 	&models.ExampleModel{},
	// 	&models.User{},
	// 	&models.Workout{},
	// 	&models.Exercise{},
	// )
	// if err != nil {
	// 	return err
	// }

	// For now, return nil since we don't have models yet
	// Uncomment and add models when you create them
	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateModels runs migrations for specific models
func MigrateModels(db *gorm.DB, models ...interface{}) error {
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
