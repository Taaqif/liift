package models

import (
	"errors"
)

// Workout represents a workout plan with multiple exercises
type Workout struct {
	BaseModel
	Name        string            `gorm:"type:varchar(255);not null" json:"name"`
	Description string            `gorm:"type:text" json:"description"`
	Exercises   []WorkoutExercise `gorm:"foreignKey:WorkoutID;constraint:OnDelete:CASCADE" json:"exercises"`
}

func (Workout) TableName() string {
	return "workouts"
}

func (w *Workout) Validate() error {
	if w.Name == "" {
		return errors.New("name_required")
	}
	return nil
}
