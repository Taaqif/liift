package repository

import (
	"context"
	"errors"
	"time"

	"liift/internal/models"
	"liift/internal/utils"

	"gorm.io/gorm"
)

var ErrActiveSessionExists = errors.New("user already has an active workout session")

type WorkoutSessionRepository struct {
	*BaseRepository
}

func NewWorkoutSessionRepository(db *gorm.DB) *WorkoutSessionRepository {
	return &WorkoutSessionRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *WorkoutSessionRepository) Start(ctx context.Context, userID uint, workoutID uint) (*models.WorkoutSession, error) {
	var active models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("user_id = ? AND ended_at IS NULL", userID).First(&active).Error
	if err == nil {
		return nil, ErrActiveSessionExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	workout, err := r.workoutByID(ctx, workoutID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session := &models.WorkoutSession{
		UserID:    userID,
		WorkoutID: workoutID,
		StartedAt: now,
	}
	if err := r.DB().WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	for i, we := range workout.Exercises {
		se := models.WorkoutSessionExercise{
			WorkoutSessionID:  session.ID,
			WorkoutExerciseID: we.ID,
			RestTimer:         we.RestTimer,
			Order:             we.Order,
		}
		if err := r.DB().WithContext(ctx).Create(&se).Error; err != nil {
			return nil, err
		}
		for j, ws := range we.Sets {
			workoutSetID := ws.ID
			ss := models.WorkoutSessionSet{
				WorkoutSessionExerciseID: se.ID,
				WorkoutSetID:             &workoutSetID,
				Order:                    j,
			}
			if err := r.DB().WithContext(ctx).Create(&ss).Error; err != nil {
				return nil, err
			}
			for _, f := range ws.Features {
				sv := models.WorkoutSessionSetValue{
					WorkoutSessionSetID: ss.ID,
					FeatureName:         f.FeatureName,
					Value:               f.Value,
				}
				if err := r.DB().WithContext(ctx).Create(&sv).Error; err != nil {
					return nil, err
				}
			}
		}
		session.Exercises = append(session.Exercises, se)
		_ = i
	}

	return r.GetByID(ctx, session.ID, userID)
}

func (r *WorkoutSessionRepository) workoutByID(ctx context.Context, id uint) (*models.Workout, error) {
	var w models.Workout
	err := r.DB().WithContext(ctx).
		Preload("Exercises", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).
		Preload("Exercises.Sets", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).
		Preload("Exercises.Sets.Features").
		First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WorkoutSessionRepository) GetActiveByUserID(ctx context.Context, userID uint) (*models.WorkoutSession, error) {
	var s models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("user_id = ? AND ended_at IS NULL", userID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, s.ID, userID)
}

func (r *WorkoutSessionRepository) GetByID(ctx context.Context, id, userID uint) (*models.WorkoutSession, error) {
	var s models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return r.loadSessionTree(ctx, &s)
}

func (r *WorkoutSessionRepository) AddExerciseToSession(ctx context.Context, sessionID, userID, exerciseID uint, restTimer int) (*models.WorkoutSession, error) {
	session, err := r.GetByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.EndedAt != nil {
		return nil, errors.New("cannot add exercise to ended session")
	}

	var exercise models.Exercise
	if err := r.DB().WithContext(ctx).Preload("ExerciseFeatures").First(&exercise, exerciseID).Error; err != nil {
		return nil, err
	}

	maxOrder := -1
	for _, e := range session.Exercises {
		if e.Order > maxOrder {
			maxOrder = e.Order
		}
	}
	nextOrder := maxOrder + 1

	se := models.WorkoutSessionExercise{
		WorkoutSessionID:  sessionID,
		WorkoutExerciseID: 0,
		ExerciseID:        &exerciseID,
		RestTimer:         restTimer,
		Order:             nextOrder,
	}
	if err := r.DB().WithContext(ctx).Create(&se).Error; err != nil {
		return nil, err
	}

	ss := models.WorkoutSessionSet{
		WorkoutSessionExerciseID: se.ID,
		Order:                    0,
	}
	if err := r.DB().WithContext(ctx).Create(&ss).Error; err != nil {
		return nil, err
	}

	for _, ef := range exercise.ExerciseFeatures {
		sv := models.WorkoutSessionSetValue{
			WorkoutSessionSetID: ss.ID,
			FeatureName:         ef.Name,
			Value:               0,
		}
		if err := r.DB().WithContext(ctx).Create(&sv).Error; err != nil {
			return nil, err
		}
	}

	return r.GetByID(ctx, sessionID, userID)
}

func (r *WorkoutSessionRepository) loadSessionTree(ctx context.Context, s *models.WorkoutSession) (*models.WorkoutSession, error) {
	err := r.DB().WithContext(ctx).
		Preload("Workout").
		Preload("Exercises", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("Exercises.WorkoutExercise.Exercise.PrimaryMuscleGroups").
		Preload("Exercises.WorkoutExercise.Exercise.SecondaryMuscleGroups").
		Preload("Exercises.WorkoutExercise.Exercise.Equipment").
		Preload("Exercises.WorkoutExercise.Exercise.ExerciseFeatures").
		Preload("Exercises.Exercise.PrimaryMuscleGroups").
		Preload("Exercises.Exercise.SecondaryMuscleGroups").
		Preload("Exercises.Exercise.Equipment").
		Preload("Exercises.Exercise.ExerciseFeatures").
		Preload("Exercises.Sets", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("Exercises.Sets.Values").
		First(s, s.ID).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *WorkoutSessionRepository) Update(ctx context.Context, session *models.WorkoutSession, userID uint) error {
	existing, err := r.GetByID(ctx, session.ID, userID)
	if err != nil {
		return err
	}
	if existing.EndedAt != nil {
		return errors.New("cannot update ended session")
	}

	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exIDs, setIDs, valueIDs := collectSessionIDs(session)
		if err := deleteRemovedSessionData(tx, existing, exIDs, setIDs, valueIDs); err != nil {
			return err
		}
		return syncSessionTree(tx, session)
	})
}

func collectSessionIDs(s *models.WorkoutSession) (exIDs, setIDs, valueIDs utils.Set[uint]) {
	exIDs, setIDs, valueIDs = utils.NewSet[uint](0), utils.NewSet[uint](0), utils.NewSet[uint](0)
	for _, ex := range s.Exercises {
		if ex.ID != 0 {
			exIDs.Add(ex.ID)
		}
		for _, set := range ex.Sets {
			if set.ID != 0 {
				setIDs.Add(set.ID)
			}
			for _, v := range set.Values {
				if v.ID != 0 {
					valueIDs.Add(v.ID)
				}
			}
		}
	}
	return exIDs, setIDs, valueIDs
}

func deleteRemovedSessionData(tx *gorm.DB, existing *models.WorkoutSession, exIDs, setIDs, valueIDs utils.Set[uint]) error {
	for _, ex := range existing.Exercises {
		for _, set := range ex.Sets {
			for _, v := range set.Values {
				if !valueIDs.Contains(v.ID) {
					if err := tx.Delete(&models.WorkoutSessionSetValue{}, v.ID).Error; err != nil {
						return err
					}
				}
			}
			if !setIDs.Contains(set.ID) {
				if err := tx.Delete(&models.WorkoutSessionSet{}, set.ID).Error; err != nil {
					return err
				}
			}
		}
		if !exIDs.Contains(ex.ID) {
			if err := tx.Delete(&models.WorkoutSessionExercise{}, ex.ID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func syncSessionTree(tx *gorm.DB, s *models.WorkoutSession) error {
	for i := range s.Exercises {
		ex := &s.Exercises[i]
		ex.WorkoutSessionID = s.ID
		if ex.ID != 0 {
			updates := map[string]interface{}{"note": ex.Note, "sort_order": ex.Order, "rest_timer": ex.RestTimer}
			if err := tx.Model(&models.WorkoutSessionExercise{}).Where("id = ? AND workout_session_id = ?", ex.ID, s.ID).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			createEx := models.WorkoutSessionExercise{
				WorkoutSessionID:  s.ID,
				WorkoutExerciseID: ex.WorkoutExerciseID,
				Note:              ex.Note,
				RestTimer:         ex.RestTimer,
				Order:             ex.Order,
			}
			if err := tx.Create(&createEx).Error; err != nil {
				return err
			}
			ex.ID = createEx.ID
		}
		for j := range ex.Sets {
			set := &ex.Sets[j]
			set.WorkoutSessionExerciseID = ex.ID
			if set.ID != 0 {
				if err := tx.Model(&models.WorkoutSessionSet{}).Where("id = ? AND workout_session_exercise_id = ?", set.ID, ex.ID).
					Updates(map[string]interface{}{"workout_set_id": set.WorkoutSetID, "sort_order": set.Order, "completed_at": set.CompletedAt}).Error; err != nil {
					return err
				}
			} else {
				createSet := models.WorkoutSessionSet{
					WorkoutSessionExerciseID: ex.ID,
					WorkoutSetID:             set.WorkoutSetID,
					Order:                    set.Order,
					CompletedAt:              set.CompletedAt,
				}
				if err := tx.Create(&createSet).Error; err != nil {
					return err
				}
				set.ID = createSet.ID
			}
			for k := range set.Values {
				v := &set.Values[k]
				v.WorkoutSessionSetID = set.ID
				if v.ID != 0 {
					if err := tx.Model(&models.WorkoutSessionSetValue{}).Where("id = ? AND workout_session_set_id = ?", v.ID, set.ID).
						Updates(map[string]interface{}{"feature_name": v.FeatureName, "value": v.Value}).Error; err != nil {
						return err
					}
				} else {
					createVal := models.WorkoutSessionSetValue{
						WorkoutSessionSetID: set.ID,
						FeatureName:         v.FeatureName,
						Value:               v.Value,
					}
					if err := tx.Create(&createVal).Error; err != nil {
						return err
					}
					v.ID = createVal.ID
				}
			}
		}
	}
	return nil
}

func (r *WorkoutSessionRepository) End(ctx context.Context, id, userID uint) (*models.WorkoutSession, error) {
	var s models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("id = ? AND user_id = ? AND ended_at IS NULL", id, userID).First(&s).Error
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := r.DB().WithContext(ctx).Model(&s).Update("ended_at", now).Error; err != nil {
		return nil, err
	}
	s.EndedAt = &now
	return r.GetByID(ctx, s.ID, userID)
}
