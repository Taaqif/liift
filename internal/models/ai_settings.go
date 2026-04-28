package models

import "gorm.io/gorm"

// AISettings stores global AI provider configuration for the entire app instance.
type AISettings struct {
	gorm.Model
	Provider      string `gorm:"not null;default:'openai'"` // openai, anthropic, google, ollama, custom
	APIKey        string `gorm:"not null;default:''"`
	AIModel       string `gorm:"column:ai_model;not null;default:''"`
	OllamaBaseURL string `gorm:"default:''"`
	CustomBaseURL string `gorm:"default:''"`
}
