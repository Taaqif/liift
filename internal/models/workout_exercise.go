package models

// WorkoutExercise represents an exercise within a workout
type WorkoutExercise struct {
	BaseModel
	WorkoutID  uint         `gorm:"not null;index" json:"workout_id"`
	ExerciseID uint         `gorm:"not null;index" json:"exercise_id"`
	RestTimer  int          `gorm:"not null;default:0" json:"rest_timer"` // seconds
	Note       string       `gorm:"type:text" json:"note"`
	Order      int          `gorm:"column:\"order\";not null;default:0" json:"order"` // for ordering exercises in workout
	Exercise   Exercise     `gorm:"foreignKey:ExerciseID" json:"exercise"`
	Sets       []WorkoutSet `gorm:"foreignKey:WorkoutExerciseID;constraint:OnDelete:CASCADE" json:"sets"`
}

func (WorkoutExercise) TableName() string {
	return "workout_exercises"
}
