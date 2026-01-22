package database

import (
	"log"

	"liift/internal/models"

	"gorm.io/gorm"
)

func SeedEquipment(db *gorm.DB) error {
	equipmentEnums := models.AllEquipmentEnums()

	for _, eqEnum := range equipmentEnums {
		equipment := models.Equipment{
			Name: string(eqEnum),
		}
		result := db.FirstOrCreate(&equipment, models.Equipment{Name: string(eqEnum)})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			log.Printf("Created equipment: %s", equipment.Name)
		}
	}

	return nil
}

func SeedMuscleGroups(db *gorm.DB) error {
	muscleGroupEnums := models.AllMuscleGroupEnums()

	for _, mgEnum := range muscleGroupEnums {
		muscleGroup := models.MuscleGroup{
			Name: string(mgEnum),
		}
		result := db.FirstOrCreate(&muscleGroup, models.MuscleGroup{Name: string(mgEnum)})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			log.Printf("Created muscle group: %s", muscleGroup.Name)
		}
	}

	return nil
}

func SeedExerciseFeatures(db *gorm.DB) error {
	exerciseFeatureEnums := models.AllExerciseFeatureEnums()

	for _, efEnum := range exerciseFeatureEnums {
		exerciseFeature := models.ExerciseFeature{
			Name: string(efEnum),
		}
		result := db.FirstOrCreate(&exerciseFeature, models.ExerciseFeature{Name: string(efEnum)})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			log.Printf("Created exercise feature: %s", exerciseFeature.Name)
		}
	}
	return nil
}

func SeedAll(db *gorm.DB) error {
	if err := SeedEquipment(db); err != nil {
		return err
	}
	if err := SeedMuscleGroups(db); err != nil {
		return err
	}
	if err := SeedExerciseFeatures(db); err != nil {
		return err
	}
	return nil
}
