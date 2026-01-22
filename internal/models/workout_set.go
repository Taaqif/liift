package models

// WorkoutSet represents a set within a workout exercise
type WorkoutSet struct {
	BaseModel
	WorkoutExerciseID uint                `gorm:"not null;index" json:"workout_exercise_id"`
	Order             int                 `gorm:"column:\"order\";not null;default:0" json:"order"` // for ordering sets
	Features          []WorkoutSetFeature `gorm:"foreignKey:WorkoutSetID;constraint:OnDelete:CASCADE" json:"features"`
}

func (WorkoutSet) TableName() string {
	return "workout_sets"
}
