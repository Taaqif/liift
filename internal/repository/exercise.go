package repository

import (
	"context"

	"liift/internal/models"

	"gorm.io/gorm"
)

// ExerciseRepository provides database operations for exercises with many-to-many associations
type ExerciseRepository struct {
	*BaseRepository
}

func NewExerciseRepository(db *gorm.DB) *ExerciseRepository {
	return &ExerciseRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *ExerciseRepository) Create(ctx context.Context, exercise *models.Exercise) error {
	return r.DBWithContext(ctx).Create(exercise).Error
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uint) (*models.Exercise, error) {
	var exercise models.Exercise
	err := r.DBWithContext(ctx).
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		First(&exercise, id).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *ExerciseRepository) List(ctx context.Context, limit, offset int) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var total int64

	db := r.DBWithContext(ctx).Model(&models.Exercise{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		Limit(limit).
		Offset(offset).
		Find(&exercises).Error; err != nil {
		return nil, 0, err
	}

	return exercises, total, nil
}

// Update uses FullSaveAssociations to ensure many-to-many relationships are properly updated
func (r *ExerciseRepository) Update(ctx context.Context, exercise *models.Exercise) error {
	return r.DBWithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(exercise).Error
}

func (r *ExerciseRepository) Delete(ctx context.Context, id uint) error {
	return r.DBWithContext(ctx).Delete(&models.Exercise{}, id).Error
}

func (r *ExerciseRepository) FindByMuscleGroup(ctx context.Context, muscleGroupID uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := r.DBWithContext(ctx).
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		Joins("JOIN exercise_primary_muscle_groups ON exercises.id = exercise_primary_muscle_groups.exercise_id").
		Where("exercise_primary_muscle_groups.muscle_group_id = ?", muscleGroupID).
		Or("EXISTS (SELECT 1 FROM exercise_secondary_muscle_groups WHERE exercise_secondary_muscle_groups.exercise_id = exercises.id AND exercise_secondary_muscle_groups.muscle_group_id = ?)", muscleGroupID).
		Find(&exercises).Error
	return exercises, err
}

func (r *ExerciseRepository) FindByEquipment(ctx context.Context, equipmentID uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := r.DBWithContext(ctx).
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		Joins("JOIN exercise_equipment ON exercises.id = exercise_equipment.exercise_id").
		Where("exercise_equipment.equipment_id = ?", equipmentID).
		Find(&exercises).Error
	return exercises, err
}

func (r *ExerciseRepository) SearchByName(ctx context.Context, query string, limit, offset int) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var total int64

	db := r.DBWithContext(ctx).Model(&models.Exercise{}).
		Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		Limit(limit).
		Offset(offset).
		Find(&exercises).Error; err != nil {
		return nil, 0, err
	}

	return exercises, total, nil
}
