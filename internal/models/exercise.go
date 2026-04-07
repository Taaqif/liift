package models

import (
	"errors"
)

// Exercise represents a workout exercise with many-to-many relationships for muscle groups, equipment, and exercise features
type Exercise struct {
	BaseModel
	Name                  string            `gorm:"type:varchar(255);not null" json:"name"`
	Description           string            `gorm:"type:text" json:"description"`
	ImageGUID             *string           `gorm:"type:varchar(36);index" json:"image_guid"`
	Force                 *string           `gorm:"type:varchar(20)" json:"force"`
	Category              *string           `gorm:"type:varchar(20)" json:"category"`
	Instructions          []string          `gorm:"serializer:json" json:"instructions"`
	PrimaryMuscleGroups   []MuscleGroup     `gorm:"many2many:exercise_primary_muscle_groups;" json:"primary_muscle_groups"`
	SecondaryMuscleGroups []MuscleGroup     `gorm:"many2many:exercise_secondary_muscle_groups;" json:"secondary_muscle_groups"`
	Equipment             []Equipment       `gorm:"many2many:exercise_equipment;" json:"equipment"`
	ExerciseFeatures      []ExerciseFeature `gorm:"many2many:exercise_exercise_features;" json:"exercise_features"`
}

func (Exercise) TableName() string {
	return "exercises"
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

	if len(e.ExerciseFeatures) == 0 {
		return errors.New("exercise_feature_required")
	}

	return nil
}
