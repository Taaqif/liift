package repository

import (
	"context"

	"liift/internal/models"

	"gorm.io/gorm"
)

type ChatRepository struct {
	BaseRepository
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{BaseRepository: BaseRepository{db: db}}
}

// Sessions

func (r *ChatRepository) CreateSession(ctx context.Context, s *models.ChatSession) error {
	return r.DB().WithContext(ctx).Create(s).Error
}

func (r *ChatRepository) GetSessionByID(ctx context.Context, id uint, userID uint) (*models.ChatSession, error) {
	var s models.ChatSession
	err := r.DB().WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ChatRepository) GetSessionBySlug(ctx context.Context, slug string, userID uint) (*models.ChatSession, error) {
	var s models.ChatSession
	err := r.DB().WithContext(ctx).
		Where("slug = ? AND user_id = ?", slug, userID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ChatRepository) ListSessions(ctx context.Context, userID uint, limit, offset int) ([]models.ChatSession, int64, error) {
	var sessions []models.ChatSession
	var total int64

	q := r.DB().WithContext(ctx).Model(&models.ChatSession{}).Where("user_id = ?", userID)
	q.Count(&total)

	err := q.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, total, err
}

func (r *ChatRepository) UpdateSessionTitle(ctx context.Context, id uint, userID uint, title string) error {
	return r.DB().WithContext(ctx).
		Model(&models.ChatSession{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("title", title).Error
}

func (r *ChatRepository) TouchSession(ctx context.Context, id uint) error {
	return r.DB().WithContext(ctx).
		Model(&models.ChatSession{}).
		Where("id = ?", id).
		Update("updated_at", gorm.Expr("NOW()")).Error
}

func (r *ChatRepository) DeleteSession(ctx context.Context, id uint, userID uint) error {
	// Delete messages first
	r.DB().WithContext(ctx).Where("session_id = ?", id).Delete(&models.ChatMessage{})
	return r.DB().WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.ChatSession{}).Error
}

// Messages

func (r *ChatRepository) CreateMessage(ctx context.Context, m *models.ChatMessage) error {
	return r.DB().WithContext(ctx).Create(m).Error
}

func (r *ChatRepository) GetMessagesBySession(ctx context.Context, sessionID uint) ([]models.ChatMessage, error) {
	var msgs []models.ChatMessage
	err := r.DB().WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&msgs).Error
	return msgs, err
}
