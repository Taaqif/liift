package models

import "time"

type WorkoutSession struct {
	BaseModel
	UserID         uint        `gorm:"not null;index" json:"user_id"`
	WorkoutID      uint        `gorm:"not null;index" json:"workout_id"`
	PlanProgressID *uint       `gorm:"index" json:"plan_progress_id,omitempty"`
	StartedAt      time.Time   `gorm:"not null" json:"started_at"`
	EndedAt        *time.Time  `json:"ended_at"`
	Workout        Workout     `gorm:"foreignKey:WorkoutID" json:"workout,omitempty"`
	Exercises      []WorkoutSessionExercise `gorm:"foreignKey:WorkoutSessionID;constraint:OnDelete:CASCADE" json:"exercises"`
}

func (WorkoutSession) TableName() string {
	return "workout_sessions"
}

type WorkoutSessionExercise struct {
	BaseModel
	WorkoutSessionID  uint                `gorm:"not null;index" json:"workout_session_id"`
	WorkoutExerciseID uint                `gorm:"index" json:"workout_exercise_id"`
	ExerciseID        *uint               `gorm:"index" json:"exercise_id,omitempty"`
	RestTimer         int                 `gorm:"not null;default:0" json:"rest_timer"`
	Order             int                 `gorm:"column:sort_order;not null;default:0" json:"order"`
	Note              string              `gorm:"type:text" json:"note"`
	WorkoutExercise   WorkoutExercise     `gorm:"foreignKey:WorkoutExerciseID" json:"workout_exercise,omitempty"`
	Exercise          Exercise            `gorm:"foreignKey:ExerciseID" json:"exercise,omitempty"`
	Sets              []WorkoutSessionSet `gorm:"foreignKey:WorkoutSessionExerciseID;constraint:OnDelete:CASCADE" json:"sets"`
}

func (WorkoutSessionExercise) TableName() string {
	return "workout_session_exercises"
}

type WorkoutSessionSet struct {
	BaseModel
	WorkoutSessionExerciseID uint                   `gorm:"not null;index" json:"workout_session_exercise_id"`
	WorkoutSetID             *uint                  `gorm:"index" json:"workout_set_id"`
	Order                    int                    `gorm:"column:sort_order;not null;default:0" json:"order"`
	CompletedAt              *time.Time             `json:"completed_at"`
	Values                   []WorkoutSessionSetValue `gorm:"foreignKey:WorkoutSessionSetID;constraint:OnDelete:CASCADE" json:"values"`
}

func (WorkoutSessionSet) TableName() string {
	return "workout_session_sets"
}

type WorkoutSessionSetValue struct {
	BaseModel
	WorkoutSessionSetID uint    `gorm:"not null;index" json:"workout_session_set_id"`
	FeatureName         string  `gorm:"type:varchar(50);not null;index" json:"feature_name"`
	Value               float64 `gorm:"not null" json:"value"`
}

func (WorkoutSessionSetValue) TableName() string {
	return "workout_session_set_values"
}
