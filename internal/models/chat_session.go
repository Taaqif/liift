package models

// ChatSession represents a conversation thread between a user and the AI coach
type ChatSession struct {
	BaseModel
	UserID   uint          `gorm:"index;not null"`
	Slug     string        `gorm:"not null;default:''"`
	Title    string        `gorm:"not null;default:'New Chat'"`
	Messages []ChatMessage `gorm:"foreignKey:SessionID"`
}
