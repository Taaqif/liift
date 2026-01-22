package models

import (
	"fmt"
)

// ExerciseFeature represents an exercise feature (weight, rep, duration)
type ExerciseFeature struct {
	Name string `gorm:"type:varchar(50);primaryKey" json:"name"`
}

func (ExerciseFeature) TableName() string {
	return "exercise_features"
}

type ExerciseFeatureEnum string

const (
	ExerciseFeatureWeight   ExerciseFeatureEnum = "weight"
	ExerciseFeatureRep      ExerciseFeatureEnum = "rep"
	ExerciseFeatureDuration ExerciseFeatureEnum = "duration"
	ExerciseFeatureDistance ExerciseFeatureEnum = "distance"
)

func AllExerciseFeatureEnums() []ExerciseFeatureEnum {
	return []ExerciseFeatureEnum{
		ExerciseFeatureWeight,
		ExerciseFeatureRep,
		ExerciseFeatureDuration,
		ExerciseFeatureDistance,
	}
}

func IsValidExerciseFeature(name string) bool {
	for _, ef := range AllExerciseFeatureEnums() {
		if string(ef) == name {
			return true
		}
	}
	return false
}

func (ef *ExerciseFeature) Validate() error {
	if !IsValidExerciseFeature(ef.Name) {
		return fmt.Errorf("invalid exercise feature name: %s", ef.Name)
	}
	return nil
}
