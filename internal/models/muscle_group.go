package models

import (
	"fmt"

	"gorm.io/gorm"
)

// MuscleGroup represents a muscle group for exercises
type MuscleGroup struct {
	BaseModel
	Name string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
}

func (MuscleGroup) TableName() string {
	return "muscle_groups"
}

type MuscleGroupEnum string

const (
	MuscleGroupAbdominals MuscleGroupEnum = "abdominals"
	MuscleGroupCardio     MuscleGroupEnum = "cardio"
	MuscleGroupShoulders  MuscleGroupEnum = "shoulders"
	MuscleGroupChest      MuscleGroupEnum = "chest"
	MuscleGroupLowerBack  MuscleGroupEnum = "lower_back"
	MuscleGroupFullBody   MuscleGroupEnum = "full_body"
	MuscleGroupBiceps     MuscleGroupEnum = "biceps"
	MuscleGroupTriceps    MuscleGroupEnum = "triceps"
	MuscleGroupUpperBack  MuscleGroupEnum = "upper_back"
	MuscleGroupQuadriceps MuscleGroupEnum = "quadriceps"
	MuscleGroupCalves     MuscleGroupEnum = "calves"
	MuscleGroupLats       MuscleGroupEnum = "lats"
	MuscleGroupHamstrings MuscleGroupEnum = "hamstrings"
	MuscleGroupGlutes     MuscleGroupEnum = "glutes"
	MuscleGroupAbductors  MuscleGroupEnum = "abductors"
	MuscleGroupAdductors  MuscleGroupEnum = "adductors"
	MuscleGroupTraps      MuscleGroupEnum = "traps"
	MuscleGroupForearms   MuscleGroupEnum = "forearms"
	MuscleGroupNeck       MuscleGroupEnum = "neck"
	MuscleGroupOther      MuscleGroupEnum = "other"
)

func AllMuscleGroupEnums() []MuscleGroupEnum {
	return []MuscleGroupEnum{
		MuscleGroupAbdominals,
		MuscleGroupCardio,
		MuscleGroupShoulders,
		MuscleGroupChest,
		MuscleGroupLowerBack,
		MuscleGroupFullBody,
		MuscleGroupBiceps,
		MuscleGroupTriceps,
		MuscleGroupUpperBack,
		MuscleGroupQuadriceps,
		MuscleGroupCalves,
		MuscleGroupLats,
		MuscleGroupHamstrings,
		MuscleGroupGlutes,
		MuscleGroupAbductors,
		MuscleGroupAdductors,
		MuscleGroupTraps,
		MuscleGroupForearms,
		MuscleGroupNeck,
		MuscleGroupOther,
	}
}

func IsValidMuscleGroup(name string) bool {
	for _, mg := range AllMuscleGroupEnums() {
		if string(mg) == name {
			return true
		}
	}
	return false
}

func (mg *MuscleGroup) Validate() error {
	if !IsValidMuscleGroup(mg.Name) {
		return fmt.Errorf("invalid muscle group name: %s", mg.Name)
	}
	return nil
}

func (mg *MuscleGroup) BeforeCreate(tx *gorm.DB) error {
	if err := mg.Validate(); err != nil {
		return err
	}
	return nil
}

func (mg *MuscleGroup) BeforeUpdate(tx *gorm.DB) error {
	if err := mg.Validate(); err != nil {
		return err
	}
	return nil
}
