package database

import (
	"log"

	"liift/internal/models"

	"gorm.io/gorm"
)

// SeedEquipment seeds the equipment table with all valid enum values
func SeedEquipment(db *gorm.DB) error {
	equipmentEnums := models.AllEquipmentEnums()

	for _, eqEnum := range equipmentEnums {
		var equipment models.Equipment
		result := db.Where("name = ?", string(eqEnum)).First(&equipment)

		if result.Error == gorm.ErrRecordNotFound {
			equipment = models.Equipment{
				Name: string(eqEnum),
			}
			if err := db.Create(&equipment).Error; err != nil {
				return err
			}
			log.Printf("Created equipment: %s", equipment.Name)
		} else if result.Error != nil {
			return result.Error
		}
		// Equipment already exists, skip
	}

	return nil
}

// SeedMuscleGroups seeds the muscle_groups table with all valid enum values
func SeedMuscleGroups(db *gorm.DB) error {
	muscleGroupEnums := models.AllMuscleGroupEnums()

	for _, mgEnum := range muscleGroupEnums {
		var muscleGroup models.MuscleGroup
		result := db.Where("name = ?", string(mgEnum)).First(&muscleGroup)

		if result.Error == gorm.ErrRecordNotFound {
			muscleGroup = models.MuscleGroup{
				Name: string(mgEnum),
			}
			if err := db.Create(&muscleGroup).Error; err != nil {
				return err
			}
			log.Printf("Created muscle group: %s", muscleGroup.Name)
		} else if result.Error != nil {
			return result.Error
		}
		// Muscle group already exists, skip
	}

	return nil
}

// SeedAll seeds all enum tables
func SeedAll(db *gorm.DB) error {
	if err := SeedEquipment(db); err != nil {
		return err
	}
	if err := SeedMuscleGroups(db); err != nil {
		return err
	}
	return nil
}
