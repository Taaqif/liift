// Package repository provides database access patterns following the repository pattern.
package repository

import (
	"context"

	"liift/internal/models"

	"gorm.io/gorm"
)

// ExampleRepository provides methods for ExampleModel database operations
// This is an example - remove when you create your actual repositories
type ExampleRepository struct {
	*BaseRepository
}

// NewExampleRepository creates a new example repository
func NewExampleRepository(db *gorm.DB) *ExampleRepository {
	return &ExampleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new example model
func (r *ExampleRepository) Create(ctx context.Context, example *models.ExampleModel) error {
	return r.DBWithContext(ctx).Create(example).Error
}

// GetByID retrieves an example model by ID
func (r *ExampleRepository) GetByID(ctx context.Context, id uint) (*models.ExampleModel, error) {
	var example models.ExampleModel
	err := r.DBWithContext(ctx).First(&example, id).Error
	if err != nil {
		return nil, err
	}
	return &example, nil
}

// List retrieves all example models with pagination
func (r *ExampleRepository) List(ctx context.Context, limit, offset int) ([]models.ExampleModel, int64, error) {
	var examples []models.ExampleModel
	var total int64

	db := r.DBWithContext(ctx).Model(&models.ExampleModel{})

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	if err := db.Limit(limit).Offset(offset).Find(&examples).Error; err != nil {
		return nil, 0, err
	}

	return examples, total, nil
}

// Update updates an example model
func (r *ExampleRepository) Update(ctx context.Context, example *models.ExampleModel) error {
	return r.DBWithContext(ctx).Save(example).Error
}

// Delete soft deletes an example model
func (r *ExampleRepository) Delete(ctx context.Context, id uint) error {
	return r.DBWithContext(ctx).Delete(&models.ExampleModel{}, id).Error
}
