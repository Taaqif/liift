package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type PlanDayJSON struct {
	IsRest      bool   `json:"is_rest"`
	WorkoutIDs  []uint `json:"workout_ids"`
	Description string `json:"description,omitempty"`
}

type PlanWeekJSON struct {
	Days []PlanDayJSON `json:"days"`
}

type ScheduleData []PlanWeekJSON

func (s ScheduleData) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *ScheduleData) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("invalid type for ScheduleData")
	}
	if len(b) == 0 || string(b) == "null" {
		*s = nil
		return nil
	}
	return json.Unmarshal(b, s)
}

type WorkoutPlan struct {
	BaseModel
	Name           string       `gorm:"type:varchar(255);not null" json:"name"`
	Description    string       `gorm:"type:text" json:"description"`
	NumberOfWeeks  int          `gorm:"not null" json:"number_of_weeks"`
	DaysPerWeek    int          `gorm:"not null" json:"days_per_week"`
	Schedule       ScheduleData `gorm:"type:jsonb" json:"schedule"`
}

func (WorkoutPlan) TableName() string {
	return "workout_plans"
}
