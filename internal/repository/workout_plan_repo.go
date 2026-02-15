package repository

import (
	"context"

	"liift/internal/models"

	"gorm.io/gorm"
)

type WorkoutPlanRepository struct {
	*BaseRepository
}

func NewWorkoutPlanRepository(db *gorm.DB) *WorkoutPlanRepository {
	return &WorkoutPlanRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *WorkoutPlanRepository) Create(ctx context.Context, plan *models.WorkoutPlan) error {
	return r.DB().WithContext(ctx).Create(plan).Error
}

func (r *WorkoutPlanRepository) GetByID(ctx context.Context, id uint) (*models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	err := r.DB().WithContext(ctx).First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *WorkoutPlanRepository) List(ctx context.Context, limit, offset int) ([]models.WorkoutPlan, int64, error) {
	var plans []models.WorkoutPlan
	var total int64

	db := r.DB().WithContext(ctx).Model(&models.WorkoutPlan{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&plans).Error; err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

func (r *WorkoutPlanRepository) Update(ctx context.Context, plan *models.WorkoutPlan) error {
	return r.DB().WithContext(ctx).Model(&models.WorkoutPlan{}).Where("id = ?", plan.ID).
		Updates(map[string]interface{}{
			"name":            plan.Name,
			"description":    plan.Description,
			"number_of_weeks": plan.NumberOfWeeks,
			"days_per_week":   plan.DaysPerWeek,
			"schedule":        plan.Schedule,
		}).Error
}

func (r *WorkoutPlanRepository) Delete(ctx context.Context, id uint) error {
	return r.DB().WithContext(ctx).Delete(&models.WorkoutPlan{}, id).Error
}
