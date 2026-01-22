package repository

import (
	"context"

	"liift/internal/models"

	"gorm.io/gorm"
)

// ImageRepository provides database operations for images
type ImageRepository struct {
	*BaseRepository
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *ImageRepository) Create(ctx context.Context, image *models.Image) error {
	return r.DB().WithContext(ctx).Create(image).Error
}

func (r *ImageRepository) GetByGUID(ctx context.Context, guid string) (*models.Image, error) {
	var image models.Image
	err := r.DB().WithContext(ctx).
		Where("guid = ?", guid).
		First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) GetByID(ctx context.Context, id uint) (*models.Image, error) {
	var image models.Image
	err := r.DB().WithContext(ctx).First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) Delete(ctx context.Context, guid string) error {
	return r.DB().WithContext(ctx).
		Where("guid = ?", guid).
		Delete(&models.Image{}).Error
}
