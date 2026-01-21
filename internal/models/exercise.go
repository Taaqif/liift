package models

import (
	"errors"

	"gorm.io/gorm"
)

// Exercise represents a workout exercise with many-to-many relationships for muscle groups and equipment
type Exercise struct {
	BaseModel
	Name                  string        `gorm:"type:varchar(255);not null" json:"name"`
	Description           string        `gorm:"type:text" json:"description"`
	Image                 string        `gorm:"type:varchar(500)" json:"image"`
	PrimaryMuscleGroups   []MuscleGroup `gorm:"many2many:exercise_primary_muscle_groups;" json:"primary_muscle_groups"`
	SecondaryMuscleGroups []MuscleGroup `gorm:"many2many:exercise_secondary_muscle_groups;" json:"secondary_muscle_groups"`
	Equipment             []Equipment   `gorm:"many2many:exercise_equipment;" json:"equipment"`
}

func (Exercise) TableName() string {
	return "exercises"
}

func (e *Exercise) BeforeCreate(tx *gorm.DB) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return nil
}

func (e *Exercise) BeforeUpdate(tx *gorm.DB) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return nil
}

func (e *Exercise) Validate() error {
	if e.Name == "" {
		return errors.New("name_required")
	}

	if len(e.PrimaryMuscleGroups) == 0 {
		return errors.New("primary_muscle_group_required")
	}

	if len(e.Equipment) == 0 {
		return errors.New("equipment_required")
	}

	return nil
}
