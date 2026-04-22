package models

import "gorm.io/gorm"

// AISettings stores per-user AI provider configuration
type AISettings struct {
	gorm.Model
	UserID        uint   `gorm:"uniqueIndex;not null"`
	Provider      string `gorm:"not null;default:'openai'"` // openai, anthropic, google, ollama, custom
	APIKey        string `gorm:"not null;default:''"`
	AIModel       string `gorm:"column:ai_model;not null;default:''"`
	OllamaBaseURL string `gorm:"default:''"`
	CustomBaseURL string `gorm:"default:''"`
}
