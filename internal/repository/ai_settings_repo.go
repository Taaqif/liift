package repository

import (
	"context"
	"errors"

	"liift/internal/models"

	"gorm.io/gorm"
)

type AISettingsRepository struct {
	BaseRepository
}

func NewAISettingsRepository(db *gorm.DB) *AISettingsRepository {
	return &AISettingsRepository{BaseRepository: BaseRepository{db: db}}
}

func (r *AISettingsRepository) GetByUserID(ctx context.Context, userID uint) (*models.AISettings, bool, error) {
	var s models.AISettings
	err := r.DB().WithContext(ctx).Where("user_id = ?", userID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.AISettings{
				UserID:   userID,
				Provider: "openai",
				APIKey:   "",
				AIModel:  "",
			}, false, nil
		}
		return nil, false, err
	}
	return &s, true, nil
}

func (r *AISettingsRepository) Upsert(ctx context.Context, s *models.AISettings) error {
	existing := &models.AISettings{}
	err := r.DB().WithContext(ctx).Where("user_id = ?", s.UserID).First(existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.DB().WithContext(ctx).Create(s).Error
		}
		return err
	}
	s.ID = existing.ID
	// Preserve existing API key if new one is empty
	if s.APIKey == "" {
		s.APIKey = existing.APIKey
	}
	return r.DB().WithContext(ctx).Save(s).Error
}
