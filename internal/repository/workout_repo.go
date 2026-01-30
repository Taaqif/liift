package repository

import (
	"context"

	"liift/internal/models"
	"liift/internal/utils"

	"gorm.io/gorm"
)

// WorkoutRepository provides database operations for workouts
type WorkoutRepository struct {
	*BaseRepository
}

func NewWorkoutRepository(db *gorm.DB) *WorkoutRepository {
	return &WorkoutRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *WorkoutRepository) Create(ctx context.Context, workout *models.Workout) error {
	return r.DB().WithContext(ctx).Create(workout).Error
}

func (r *WorkoutRepository) GetByID(ctx context.Context, id uint) (*models.Workout, error) {
	var workout models.Workout
	err := r.DB().WithContext(ctx).
		Preload("Exercises", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Exercises.Exercise.PrimaryMuscleGroups").
		Preload("Exercises.Exercise.SecondaryMuscleGroups").
		Preload("Exercises.Exercise.Equipment").
		Preload("Exercises.Exercise.ExerciseFeatures").
		Preload("Exercises.Sets", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Exercises.Sets.Features").
		First(&workout, id).Error
	if err != nil {
		return nil, err
	}
	return &workout, nil
}

func (r *WorkoutRepository) List(
	ctx context.Context,
	limit, offset int,
	search, exerciseFeature string,
	exerciseIDs []uint,
	muscleGroups, equipment []string,
) ([]models.Workout, int64, error) {
	var workouts []models.Workout
	var total int64

	db := r.DB().WithContext(ctx).Model(&models.Workout{})

	if search != "" {
		pattern := "%" + search + "%"
		db = db.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", pattern, pattern)
	}
	if exerciseFeature != "" {
		featSubq := r.DB().WithContext(ctx).Table("workout_exercises").
			Select("1").
			Joins("JOIN exercise_exercise_features ON exercise_exercise_features.exercise_id = workout_exercises.exercise_id").
			Where("workout_exercises.workout_id = workouts.id").
			Where("exercise_exercise_features.exercise_feature_name = ?", exerciseFeature)
		db = db.Where("EXISTS (?)", featSubq)
	}
	if len(exerciseIDs) > 0 {
		exSubq := r.DB().WithContext(ctx).Table("workout_exercises").
			Select("1").
			Where("workout_exercises.workout_id = workouts.id").
			Where("workout_exercises.exercise_id IN ?", exerciseIDs)
		db = db.Where("EXISTS (?)", exSubq)
	}
	if len(muscleGroups) > 0 {
		primarySubq := r.DB().WithContext(ctx).Table("workout_exercises").
			Select("1").
			Joins("JOIN exercise_primary_muscle_groups ON exercise_primary_muscle_groups.exercise_id = workout_exercises.exercise_id").
			Where("workout_exercises.workout_id = workouts.id").
			Where("exercise_primary_muscle_groups.muscle_group_name IN ?", muscleGroups)
		secondarySubq := r.DB().WithContext(ctx).Table("workout_exercises").
			Select("1").
			Joins("JOIN exercise_secondary_muscle_groups ON exercise_secondary_muscle_groups.exercise_id = workout_exercises.exercise_id").
			Where("workout_exercises.workout_id = workouts.id").
			Where("exercise_secondary_muscle_groups.muscle_group_name IN ?", muscleGroups)
		db = db.Where("EXISTS (?) OR EXISTS (?)", primarySubq, secondarySubq)
	}
	if len(equipment) > 0 {
		equipSubq := r.DB().WithContext(ctx).Table("workout_exercises").
			Select("1").
			Joins("JOIN exercise_equipment ON exercise_equipment.exercise_id = workout_exercises.exercise_id").
			Where("workout_exercises.workout_id = workouts.id").
			Where("exercise_equipment.equipment_name IN ?", equipment)
		db = db.Where("EXISTS (?)", equipSubq)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Preload("Exercises", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Exercises.Exercise.PrimaryMuscleGroups").
		Preload("Exercises.Exercise.SecondaryMuscleGroups").
		Preload("Exercises.Exercise.Equipment").
		Preload("Exercises.Exercise.ExerciseFeatures").
		Preload("Exercises.Sets", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Exercises.Sets.Features").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&workouts).Error; err != nil {
		return nil, 0, err
	}

	return workouts, total, nil
}

func (r *WorkoutRepository) Update(ctx context.Context, workout *models.Workout) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Workout{}).Where("id = ?", workout.ID).
			Updates(map[string]interface{}{"name": workout.Name, "description": workout.Description}).Error; err != nil {
			return err
		}
		var existing models.Workout
		if err := tx.Where("id = ?", workout.ID).Preload("Exercises").Preload("Exercises.Sets").Preload("Exercises.Sets.Features").
			First(&existing).Error; err != nil {
			return err
		}
		exIDs, setIDs, featIDs := collectIncomingIDs(workout)
		if err := deleteRemoved(tx, &existing, exIDs, setIDs, featIDs); err != nil {
			return err
		}
		return syncWorkoutTree(tx, workout)
	})
}

func collectIncomingIDs(w *models.Workout) (exIDs, setIDs, featIDs utils.Set[uint]) {
	exIDs, setIDs, featIDs = utils.NewSet[uint](0), utils.NewSet[uint](0), utils.NewSet[uint](0)
	for _, ex := range w.Exercises {
		if ex.ID != 0 {
			exIDs.Add(ex.ID)
		}
		for _, set := range ex.Sets {
			if set.ID != 0 {
				setIDs.Add(set.ID)
			}
			for _, f := range set.Features {
				if f.ID != 0 {
					featIDs.Add(f.ID)
				}
			}
		}
	}
	return exIDs, setIDs, featIDs
}

func deleteRemoved(tx *gorm.DB, existing *models.Workout, exIDs, setIDs, featIDs utils.Set[uint]) error {
	for _, ex := range existing.Exercises {
		for _, set := range ex.Sets {
			for _, f := range set.Features {
				if !featIDs.Contains(f.ID) {
					if err := tx.Delete(&models.WorkoutSetFeature{}, f.ID).Error; err != nil {
						return err
					}
				}
			}
			if !setIDs.Contains(set.ID) {
				if err := tx.Delete(&models.WorkoutSet{}, set.ID).Error; err != nil {
					return err
				}
			}
		}
		if !exIDs.Contains(ex.ID) {
			if err := tx.Delete(&models.WorkoutExercise{}, ex.ID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func syncWorkoutTree(tx *gorm.DB, w *models.Workout) error {
	for i := range w.Exercises {
		ex := &w.Exercises[i]
		ex.WorkoutID = w.ID
		if ex.ID != 0 {
			if err := tx.Model(&models.WorkoutExercise{}).Where("id = ? AND workout_id = ?", ex.ID, w.ID).
				Updates(map[string]interface{}{"exercise_id": ex.ExerciseID, "rest_timer": ex.RestTimer, "note": ex.Note, "order": ex.Order}).Error; err != nil {
				return err
			}
		} else {
			createEx := models.WorkoutExercise{WorkoutID: w.ID, ExerciseID: ex.ExerciseID, RestTimer: ex.RestTimer, Note: ex.Note, Order: ex.Order}
			if err := tx.Create(&createEx).Error; err != nil {
				return err
			}
			ex.ID = createEx.ID
		}
		for j := range ex.Sets {
			set := &ex.Sets[j]
			set.WorkoutExerciseID = ex.ID
			if set.ID != 0 {
				if err := tx.Model(&models.WorkoutSet{}).Where("id = ? AND workout_exercise_id = ?", set.ID, ex.ID).Updates(map[string]interface{}{"order": set.Order}).Error; err != nil {
					return err
				}
			} else {
				createSet := models.WorkoutSet{WorkoutExerciseID: ex.ID, Order: set.Order}
				if err := tx.Create(&createSet).Error; err != nil {
					return err
				}
				set.ID = createSet.ID
			}
			for k := range set.Features {
				feat := &set.Features[k]
				feat.WorkoutSetID = set.ID
				if feat.ID != 0 {
					if err := tx.Model(&models.WorkoutSetFeature{}).Where("id = ? AND workout_set_id = ?", feat.ID, set.ID).
						Updates(map[string]interface{}{"feature_name": feat.FeatureName, "value": feat.Value}).Error; err != nil {
						return err
					}
				} else {
					createFeat := models.WorkoutSetFeature{WorkoutSetID: set.ID, FeatureName: feat.FeatureName, Value: feat.Value}
					if err := tx.Create(&createFeat).Error; err != nil {
						return err
					}
					feat.ID = createFeat.ID
				}
			}
		}
	}
	return nil
}

func (r *WorkoutRepository) Delete(ctx context.Context, id uint) error {
	// Workout has constraint:OnDelete:CASCADE on Exercises; DB cascades to sets and features
	return r.DB().WithContext(ctx).Delete(&models.Workout{}, id).Error
}
