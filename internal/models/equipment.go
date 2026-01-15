package models

import (
	"fmt"

	"gorm.io/gorm"
)

// Equipment represents a piece of exercise equipment
type Equipment struct {
	Name string `gorm:"type:varchar(50);primaryKey" json:"name"`
}

func (Equipment) TableName() string {
	return "equipment"
}

type EquipmentEnum string

const (
	EquipmentBarbell        EquipmentEnum = "barbell"
	EquipmentBench          EquipmentEnum = "bench"
	EquipmentInclineBench   EquipmentEnum = "incline_bench"
	EquipmentDumbbell       EquipmentEnum = "dumbbell"
	EquipmentKettlebell     EquipmentEnum = "kettlebell"
	EquipmentMachine        EquipmentEnum = "machine"
	EquipmentPlate          EquipmentEnum = "plate"
	EquipmentResistanceBand EquipmentEnum = "resistance_band"
	EquipmentGymMat         EquipmentEnum = "gym_mat"
	EquipmentPullUpBar      EquipmentEnum = "pull_up_bar"
	EquipmentSwissBall      EquipmentEnum = "swiss_ball"
	EquipmentBodyweight     EquipmentEnum = "bodyweight"
)

func AllEquipmentEnums() []EquipmentEnum {
	return []EquipmentEnum{
		EquipmentBarbell,
		EquipmentBench,
		EquipmentInclineBench,
		EquipmentDumbbell,
		EquipmentKettlebell,
		EquipmentMachine,
		EquipmentPlate,
		EquipmentResistanceBand,
		EquipmentGymMat,
		EquipmentPullUpBar,
		EquipmentSwissBall,
		EquipmentBodyweight,
	}
}

func IsValidEquipment(name string) bool {
	for _, e := range AllEquipmentEnums() {
		if string(e) == name {
			return true
		}
	}
	return false
}

func (e *Equipment) Validate() error {
	if !IsValidEquipment(e.Name) {
		return fmt.Errorf("invalid equipment name: %s", e.Name)
	}
	return nil
}

func (e *Equipment) BeforeCreate(tx *gorm.DB) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return nil
}

func (e *Equipment) BeforeUpdate(tx *gorm.DB) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return nil
}
