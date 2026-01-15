// Package models contains database models for the application.
package models

// ExampleModel is an example model demonstrating GORM usage
// Remove this file when you create your actual models
type ExampleModel struct {
	BaseModel
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for this model
func (ExampleModel) TableName() string {
	return "example_models"
}
