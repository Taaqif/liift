package repository

import (
	"context"
	"errors"
	"time"

	"liift/internal/models"

	"gorm.io/gorm"
)

var ErrActivePlanProgressExists = errors.New("active plan progress already exists")

type WorkoutPlanProgressRepository struct {
	*BaseRepository
}

func NewWorkoutPlanProgressRepository(db *gorm.DB) *WorkoutPlanProgressRepository {
	return &WorkoutPlanProgressRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *WorkoutPlanProgressRepository) GetActiveByUserID(ctx context.Context, userID uint) (*models.WorkoutPlanProgress, error) {
	var progress models.WorkoutPlanProgress
	err := r.DB().WithContext(ctx).
		Where("user_id = ? AND completed_at IS NULL", userID).
		Preload("Plan").
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *WorkoutPlanProgressRepository) GetByID(ctx context.Context, id, userID uint) (*models.WorkoutPlanProgress, error) {
	var progress models.WorkoutPlanProgress
	q := r.DB().WithContext(ctx).Where("id = ?", id)
	if userID != 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.Preload("Plan").First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *WorkoutPlanProgressRepository) Start(ctx context.Context, userID, planID uint) (*models.WorkoutPlanProgress, error) {
	// Only one active plan at a time
	var existing models.WorkoutPlanProgress
	err := r.DB().WithContext(ctx).
		Where("user_id = ? AND completed_at IS NULL", userID).
		First(&existing).Error
	if err == nil {
		return nil, ErrActivePlanProgressExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	progress := &models.WorkoutPlanProgress{
		UserID:      userID,
		PlanID:      planID,
		CurrentWeek: 0,
		CurrentDay:  0,
		StartedAt:   time.Now().UTC(),
	}

	if err := r.DB().WithContext(ctx).Create(progress).Error; err != nil {
		return nil, err
	}

	return r.GetByID(ctx, progress.ID, userID)
}

func (r *WorkoutPlanProgressRepository) UpdatePosition(ctx context.Context, id, userID uint, week, day int) (*models.WorkoutPlanProgress, error) {
	if err := r.DB().WithContext(ctx).
		Model(&models.WorkoutPlanProgress{}).
		Where("id = ? AND user_id = ? AND completed_at IS NULL", id, userID).
		Updates(map[string]interface{}{
			"current_week": week,
			"current_day":  day,
		}).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, userID)
}

func (r *WorkoutPlanProgressRepository) Complete(ctx context.Context, id, userID uint) (*models.WorkoutPlanProgress, error) {
	now := time.Now().UTC()
	if err := r.DB().WithContext(ctx).
		Model(&models.WorkoutPlanProgress{}).
		Where("id = ? AND user_id = ? AND completed_at IS NULL", id, userID).
		Update("completed_at", now).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, userID)
}

// AdvanceDay moves the plan progress to the next day.
// If already at the last day, it's a no-op (user must manually complete).
func (r *WorkoutPlanProgressRepository) AdvanceDay(ctx context.Context, id uint) error {
	progress, err := r.GetByID(ctx, id, 0 /* no userID check — internal call */)
	if err != nil {
		return err
	}
	if progress.CompletedAt != nil {
		return nil
	}

	schedule := progress.Plan.Schedule
	// Find next day after current position
	for w := progress.CurrentWeek; w < len(schedule); w++ {
		startDay := 0
		if w == progress.CurrentWeek {
			startDay = progress.CurrentDay + 1
		}
		for d := startDay; d < len(schedule[w].Days); d++ {
			if len(schedule[w].Days[d].WorkoutIDs) == 0 {
				continue // skip rest days
			}
			return r.DB().WithContext(ctx).
				Model(&models.WorkoutPlanProgress{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"current_week": w,
					"current_day":  d,
				}).Error
		}
	}
	// No next day found — at end of plan, nothing to do
	return nil
}

func (r *WorkoutPlanProgressRepository) Stop(ctx context.Context, id, userID uint) error {
	return r.DB().WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.WorkoutPlanProgress{}).Error
}
