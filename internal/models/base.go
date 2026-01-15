// Package models contains database models for the application.
package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// BeforeCreate is a GORM hook that runs before creating a record
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	// Ensure timestamps are set to UTC
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now().UTC()
	return nil
}
