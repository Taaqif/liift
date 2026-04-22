package models

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	BaseModel
	SessionID uint   `gorm:"index;not null"`
	Role      string `gorm:"not null"` // user, assistant
	Content   string `gorm:"not null;type:text"`
	// JSON string storing artifact data (workout/plan) attached to this message
	Metadata string `gorm:"type:text;default:''"`
}
