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
	return r.DB().WithContext(ctx).Create(exercise).Error
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uint) (*models.Exercise, error) {
	var exercise models.Exercise
	err := r.DB().WithContext(ctx).
		Preload("PrimaryMuscleGroups").
		Preload("SecondaryMuscleGroups").
		Preload("Equipment").
		First(&exercise, id).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *ExerciseRepository) List(
	ctx context.Context,
	limit, offset int,
	search string,
	muscleGroups []string,
	equipment []string,
) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var total int64

	db := r.DB().WithContext(ctx).Model(&models.Exercise{})

	if search != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	if len(muscleGroups) > 0 {
		primarySubquery := r.DB().WithContext(ctx).
			Table("exercise_primary_muscle_groups").
			Select("1").
			Where("exercise_primary_muscle_groups.exercise_id = exercises.id").
			Where("muscle_group_name IN ?", muscleGroups)

		secondarySubquery := r.DB().WithContext(ctx).
			Table("exercise_secondary_muscle_groups").
			Select("1").
			Where("exercise_secondary_muscle_groups.exercise_id = exercises.id").
			Where("muscle_group_name IN ?", muscleGroups)

		db = db.Where("EXISTS (?) OR EXISTS (?)", primarySubquery, secondarySubquery)
	}

	if len(equipment) > 0 {
		equipmentSubquery := r.DB().WithContext(ctx).
			Table("exercise_equipment").
			Select("1").
			Where("exercise_equipment.exercise_id = exercises.id").
			Where("equipment_name IN ?", equipment)

		db = db.Where("EXISTS (?)", equipmentSubquery)
	}

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

func (r *ExerciseRepository) Update(ctx context.Context, exercise *models.Exercise) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Exercise{}).Where("id = ?", exercise.ID).
			Updates(models.Exercise{
				Name:        exercise.Name,
				Description: exercise.Description,
				Image:       exercise.Image,
			}).Error; err != nil {
			return err
		}

		target := models.Exercise{BaseModel: models.BaseModel{ID: exercise.ID}}

		if err := tx.Model(&target).Association("Equipment").Clear(); err != nil {
			return err
		}
		if len(exercise.Equipment) > 0 {
			if err := tx.Model(&target).Association("Equipment").Append(exercise.Equipment); err != nil {
				return err
			}
		}

		if err := tx.Model(&target).Association("PrimaryMuscleGroups").Clear(); err != nil {
			return err
		}
		if len(exercise.PrimaryMuscleGroups) > 0 {
			if err := tx.Model(&target).Association("PrimaryMuscleGroups").Append(exercise.PrimaryMuscleGroups); err != nil {
				return err
			}
		}

		if err := tx.Model(&target).Association("SecondaryMuscleGroups").Clear(); err != nil {
			return err
		}
		if len(exercise.SecondaryMuscleGroups) > 0 {
			if err := tx.Model(&target).Association("SecondaryMuscleGroups").Append(exercise.SecondaryMuscleGroups); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ExerciseRepository) Delete(ctx context.Context, id uint) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		target := models.Exercise{BaseModel: models.BaseModel{ID: id}}

		if err := tx.Model(&target).Association("Equipment").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&target).Association("PrimaryMuscleGroups").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&target).Association("SecondaryMuscleGroups").Clear(); err != nil {
			return err
		}

		return tx.Delete(&models.Exercise{}, id).Error
	})
}
