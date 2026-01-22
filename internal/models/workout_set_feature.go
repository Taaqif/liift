package models

// WorkoutSetFeature represents a feature value for a workout set
type WorkoutSetFeature struct {
	BaseModel
	WorkoutSetID uint    `gorm:"not null;index" json:"workout_set_id"`
	FeatureName  string  `gorm:"type:varchar(50);not null;index" json:"feature_name"`
	Value        float64 `gorm:"not null" json:"value"`
}

func (WorkoutSetFeature) TableName() string {
	return "workout_set_features"
}
