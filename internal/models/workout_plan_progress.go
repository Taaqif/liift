package models

import "time"

type WorkoutPlanProgress struct {
	BaseModel
	UserID      uint        `gorm:"not null;index" json:"user_id"`
	PlanID      uint        `gorm:"not null" json:"plan_id"`
	Plan        WorkoutPlan `json:"plan"`
	CurrentWeek int         `gorm:"not null;default:0" json:"current_week"`
	CurrentDay  int         `gorm:"not null;default:0" json:"current_day"`
	StartedAt   time.Time   `gorm:"not null" json:"started_at"`
	CompletedAt *time.Time  `json:"completed_at"`
}

func (WorkoutPlanProgress) TableName() string {
	return "workout_plan_progress"
}
