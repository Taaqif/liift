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

// lastExerciseValues returns the most recent completed set value per (exerciseID, featureName)
// for the given user and exercise IDs. Only non-zero values from completed sets are returned.
func (r *WorkoutSessionRepository) lastExerciseValues(ctx context.Context, userID uint, exerciseIDs []uint) (map[uint]map[string]float64, error) {
	if len(exerciseIDs) == 0 {
		return nil, nil
	}

	type row struct {
		ExerciseID  uint
		FeatureName string
		Value       float64
		CompletedAt time.Time
	}
	var rows []row
	if err := r.DB().WithContext(ctx).Raw(`
		SELECT
			COALESCE(we.exercise_id, wse.exercise_id) AS exercise_id,
			wssv.feature_name,
			wssv.value,
			wss.completed_at
		FROM workout_session_set_values wssv
		JOIN workout_session_sets wss
			ON wss.id = wssv.workout_session_set_id
			AND wss.deleted_at IS NULL
			AND wss.completed_at IS NOT NULL
		JOIN workout_session_exercises wse
			ON wse.id = wss.workout_session_exercise_id
			AND wse.deleted_at IS NULL
		LEFT JOIN workout_exercises we
			ON we.id = wse.workout_exercise_id
			AND we.deleted_at IS NULL
		JOIN workout_sessions ws
			ON ws.id = wse.workout_session_id
			AND ws.deleted_at IS NULL
		WHERE ws.user_id = ?
		  AND wssv.value != 0
		  AND wssv.deleted_at IS NULL
		  AND COALESCE(we.exercise_id, wse.exercise_id) IN ?
		ORDER BY wss.completed_at DESC
	`, userID, exerciseIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]map[string]float64)
	for _, row := range rows {
		if _, ok := result[row.ExerciseID]; !ok {
			result[row.ExerciseID] = make(map[string]float64)
		}
		// First row per (exerciseID, featureName) is the most recent
		if _, seen := result[row.ExerciseID][row.FeatureName]; !seen {
			result[row.ExerciseID][row.FeatureName] = row.Value
		}
	}
	return result, nil
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

	exerciseIDs := make([]uint, 0, len(workout.Exercises))
	for _, we := range workout.Exercises {
		exerciseIDs = append(exerciseIDs, we.ExerciseID)
	}
	lastValues, err := r.lastExerciseValues(ctx, userID, exerciseIDs)
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
				val := f.Value
				if val == 0 {
					if last, ok := lastValues[we.ExerciseID][f.FeatureName]; ok {
						val = last
					}
				}
				sv := models.WorkoutSessionSetValue{
					WorkoutSessionSetID: ss.ID,
					FeatureName:         f.FeatureName,
					Value:               val,
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

func (r *WorkoutSessionRepository) StartBlank(ctx context.Context, userID uint, name string) (*models.WorkoutSession, error) {
	var active models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("user_id = ? AND ended_at IS NULL", userID).First(&active).Error
	if err == nil {
		return nil, ErrActiveSessionExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	workout := &models.Workout{
		Name:      name,
		IsLibrary: false,
		IsManual:  true,
	}
	if err := r.DB().WithContext(ctx).Create(workout).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session := &models.WorkoutSession{
		UserID:    userID,
		WorkoutID: workout.ID,
		StartedAt: now,
	}
	if err := r.DB().WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	return r.GetByID(ctx, session.ID, userID)
}

func (r *WorkoutSessionRepository) StartDay(ctx context.Context, userID, planProgressID uint, workoutIDs []uint) (*models.WorkoutSession, error) {
	var active models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("user_id = ? AND ended_at IS NULL", userID).First(&active).Error
	if err == nil {
		return nil, ErrActiveSessionExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if len(workoutIDs) == 0 {
		return nil, errors.New("no workouts provided")
	}

	// Collect all exercise IDs across all workouts for last-value lookup
	workouts := make([]*models.Workout, 0, len(workoutIDs))
	allExerciseIDs := make([]uint, 0)
	for _, wid := range workoutIDs {
		workout, err := r.workoutByID(ctx, wid)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
		for _, we := range workout.Exercises {
			allExerciseIDs = append(allExerciseIDs, we.ExerciseID)
		}
	}
	lastValues, err := r.lastExerciseValues(ctx, userID, allExerciseIDs)
	if err != nil {
		return nil, err
	}

	primaryWorkoutID := workoutIDs[0]

	now := time.Now().UTC()
	ppID := planProgressID
	session := &models.WorkoutSession{
		UserID:         userID,
		WorkoutID:      primaryWorkoutID,
		PlanProgressID: &ppID,
		StartedAt:      now,
	}
	if err := r.DB().WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	order := 0
	for _, workout := range workouts {
		for _, we := range workout.Exercises {
			se := models.WorkoutSessionExercise{
				WorkoutSessionID:  session.ID,
				WorkoutExerciseID: we.ID,
				RestTimer:         we.RestTimer,
				Order:             order,
			}
			order++
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
					val := f.Value
					if val == 0 {
						if last, ok := lastValues[we.ExerciseID][f.FeatureName]; ok {
							val = last
						}
					}
					sv := models.WorkoutSessionSetValue{
						WorkoutSessionSetID: ss.ID,
						FeatureName:         f.FeatureName,
						Value:               val,
					}
					if err := r.DB().WithContext(ctx).Create(&sv).Error; err != nil {
						return nil, err
					}
				}
			}
		}
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

type ExerciseLogValue struct {
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type ExerciseLogSet struct {
	Order       int                `json:"order"`
	CompletedAt time.Time          `json:"completed_at"`
	Values      []ExerciseLogValue `json:"values"`
}

type ExerciseLogEntry struct {
	SessionID   uint             `json:"session_id"`
	Date        time.Time        `json:"date"`
	WorkoutName string           `json:"workout_name"`
	Sets        []ExerciseLogSet `json:"sets"`
}

func (r *WorkoutSessionRepository) GetExerciseLogs(ctx context.Context, exerciseID uint, userID uint, from, to *time.Time, limit, offset int) ([]ExerciseLogEntry, int64, error) {
	dateFilter := ""
	var dateArgs []interface{}
	if from != nil {
		dayStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		dateFilter += " AND ws.started_at >= ?"
		dateArgs = append(dateArgs, dayStart)
	}
	if to != nil {
		dayEnd := time.Date(to.Year(), to.Month(), to.Day()+1, 0, 0, 0, 0, time.UTC)
		dateFilter += " AND ws.started_at < ?"
		dateArgs = append(dateArgs, dayEnd)
	}

	type countResult struct {
		Count int64
	}
	var cr countResult
	countArgs := append([]interface{}{exerciseID, exerciseID, userID}, dateArgs...)
	if err := r.DB().WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT ws.id) as count
		FROM workout_sessions ws
		JOIN workout_session_exercises wse ON wse.workout_session_id = ws.id AND wse.deleted_at IS NULL
		LEFT JOIN workout_exercises we ON we.id = wse.workout_exercise_id AND we.deleted_at IS NULL
		JOIN workout_session_sets wss ON wss.workout_session_exercise_id = wse.id AND wss.deleted_at IS NULL
		WHERE (we.exercise_id = ? OR wse.exercise_id = ?)
		AND wss.completed_at IS NOT NULL
		AND ws.deleted_at IS NULL
		AND ws.user_id = ?`+dateFilter+`
	`, countArgs...).Scan(&cr).Error; err != nil {
		return nil, 0, err
	}
	if cr.Count == 0 {
		return nil, 0, nil
	}

	type sessionMeta struct {
		ID          uint
		StartedAt   time.Time
		WorkoutName string
	}
	var metas []sessionMeta
	metaArgs := append([]interface{}{exerciseID, exerciseID, userID}, dateArgs...)
	metaArgs = append(metaArgs, limit, offset)
	if err := r.DB().WithContext(ctx).Raw(`
		SELECT DISTINCT ws.id, ws.started_at, COALESCE(w.name, '') as workout_name
		FROM workout_sessions ws
		JOIN workout_session_exercises wse ON wse.workout_session_id = ws.id AND wse.deleted_at IS NULL
		LEFT JOIN workout_exercises we ON we.id = wse.workout_exercise_id AND we.deleted_at IS NULL
		LEFT JOIN workouts w ON w.id = ws.workout_id AND w.deleted_at IS NULL
		JOIN workout_session_sets wss ON wss.workout_session_exercise_id = wse.id AND wss.deleted_at IS NULL
		WHERE (we.exercise_id = ? OR wse.exercise_id = ?)
		AND wss.completed_at IS NOT NULL
		AND ws.deleted_at IS NULL
		AND ws.user_id = ?`+dateFilter+`
		ORDER BY ws.started_at DESC
		LIMIT ? OFFSET ?
	`, metaArgs...).Scan(&metas).Error; err != nil {
		return nil, 0, err
	}
	if len(metas) == 0 {
		return nil, cr.Count, nil
	}

	sessionIDs := make([]uint, len(metas))
	for i, m := range metas {
		sessionIDs[i] = m.ID
	}

	type setRow struct {
		SessionID   uint
		SetID       uint
		SetOrder    int
		CompletedAt time.Time
		FeatureName string
		Value       float64
	}
	var rows []setRow
	if err := r.DB().WithContext(ctx).Raw(`
		SELECT wse.workout_session_id as session_id,
		       wss.id as set_id,
		       wss.sort_order as set_order,
		       wss.completed_at,
		       wssv.feature_name,
		       wssv.value
		FROM workout_session_sets wss
		JOIN workout_session_exercises wse ON wse.id = wss.workout_session_exercise_id AND wse.deleted_at IS NULL
		LEFT JOIN workout_exercises we ON we.id = wse.workout_exercise_id AND we.deleted_at IS NULL
		JOIN workout_session_set_values wssv ON wssv.workout_session_set_id = wss.id AND wssv.deleted_at IS NULL
		WHERE wse.workout_session_id IN ?
		AND (we.exercise_id = ? OR wse.exercise_id = ?)
		AND wss.completed_at IS NOT NULL
		AND wss.deleted_at IS NULL
		ORDER BY wse.sort_order ASC, wss.sort_order ASC
	`, sessionIDs, exerciseID, exerciseID).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	type setInfo struct {
		order       int
		completedAt time.Time
		values      []ExerciseLogValue
	}
	sessionSetIDs := map[uint][]uint{}
	setInfos := map[uint]*setInfo{}
	for _, row := range rows {
		if _, exists := setInfos[row.SetID]; !exists {
			setInfos[row.SetID] = &setInfo{order: row.SetOrder, completedAt: row.CompletedAt}
			sessionSetIDs[row.SessionID] = append(sessionSetIDs[row.SessionID], row.SetID)
		}
		setInfos[row.SetID].values = append(setInfos[row.SetID].values, ExerciseLogValue{
			FeatureName: row.FeatureName,
			Value:       row.Value,
		})
	}

	result := make([]ExerciseLogEntry, 0, len(metas))
	for _, meta := range metas {
		var logSets []ExerciseLogSet
		for _, setID := range sessionSetIDs[meta.ID] {
			s := setInfos[setID]
			logSets = append(logSets, ExerciseLogSet{
				Order:       s.order,
				CompletedAt: s.completedAt,
				Values:      s.values,
			})
		}
		if len(logSets) == 0 {
			continue
		}
		result = append(result, ExerciseLogEntry{
			SessionID:   meta.ID,
			Date:        meta.StartedAt,
			WorkoutName: meta.WorkoutName,
			Sets:        logSets,
		})
	}

	return result, cr.Count, nil
}

type WorkoutSessionSummary struct {
	ID            uint       `json:"id"`
	WorkoutID     uint       `json:"workout_id"`
	WorkoutName   string     `json:"workout_name"`
	StartedAt     time.Time  `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at"`
	ExerciseCount int        `json:"exercise_count"`
	SetsCompleted int        `json:"sets_completed"`
}

func (r *WorkoutSessionRepository) ListByUserID(ctx context.Context, userID uint, workoutID *uint, date *time.Time, from, to *time.Time, limit, offset int) ([]WorkoutSessionSummary, int64, error) {
	db := r.DB().WithContext(ctx).Model(&models.WorkoutSession{}).
		Where("user_id = ? AND ended_at IS NOT NULL AND deleted_at IS NULL", userID)
	if workoutID != nil {
		db = db.Where("workout_id = ?", *workoutID)
	}
	if date != nil {
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		dayEnd := dayStart.AddDate(0, 0, 1)
		db = db.Where("started_at >= ? AND started_at < ?", dayStart, dayEnd)
	}
	if from != nil {
		dayStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		db = db.Where("started_at >= ?", dayStart)
	}
	if to != nil {
		dayEnd := time.Date(to.Year(), to.Month(), to.Day()+1, 0, 0, 0, 0, time.UTC)
		db = db.Where("started_at < ?", dayEnd)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	extraFilters := ""
	args := []interface{}{userID}
	if workoutID != nil {
		extraFilters += " AND ws.workout_id = ?"
		args = append(args, *workoutID)
	}
	if date != nil {
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		dayEnd := dayStart.AddDate(0, 0, 1)
		extraFilters += " AND ws.started_at >= ? AND ws.started_at < ?"
		args = append(args, dayStart, dayEnd)
	}
	if from != nil {
		dayStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		extraFilters += " AND ws.started_at >= ?"
		args = append(args, dayStart)
	}
	if to != nil {
		dayEnd := time.Date(to.Year(), to.Month(), to.Day()+1, 0, 0, 0, 0, time.UTC)
		extraFilters += " AND ws.started_at < ?"
		args = append(args, dayEnd)
	}
	args = append(args, limit, offset)

	var rows []WorkoutSessionSummary
	if err := r.DB().WithContext(ctx).Raw(`
		SELECT
			ws.id,
			ws.workout_id,
			COALESCE(w.name, '') AS workout_name,
			ws.started_at,
			ws.ended_at,
			COUNT(DISTINCT wse.id) AS exercise_count,
			COUNT(DISTINCT CASE WHEN wss.completed_at IS NOT NULL THEN wss.id END) AS sets_completed
		FROM workout_sessions ws
		LEFT JOIN workouts w ON w.id = ws.workout_id AND w.deleted_at IS NULL
		LEFT JOIN workout_session_exercises wse ON wse.workout_session_id = ws.id AND wse.deleted_at IS NULL
		LEFT JOIN workout_session_sets wss ON wss.workout_session_exercise_id = wse.id AND wss.deleted_at IS NULL
		WHERE ws.user_id = ? AND ws.ended_at IS NOT NULL AND ws.deleted_at IS NULL`+extraFilters+`
		GROUP BY ws.id, w.name, ws.workout_id, ws.started_at, ws.ended_at
		ORDER BY ws.started_at DESC
		LIMIT ? OFFSET ?
	`, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// ListActivityDates returns distinct "YYYY-MM-DD" dates that have completed sessions
// for the given user in the given year/month.
func (r *WorkoutSessionRepository) ListActivityDates(ctx context.Context, userID uint, year, month int) ([]string, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	var sessions []struct {
		StartedAt time.Time
	}
	if err := r.DB().WithContext(ctx).
		Model(&models.WorkoutSession{}).
		Select("started_at").
		Where("user_id = ? AND ended_at IS NOT NULL AND deleted_at IS NULL AND started_at >= ? AND started_at < ?", userID, start, end).
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var dates []string
	for _, s := range sessions {
		key := s.StartedAt.UTC().Format("2006-01-02")
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			dates = append(dates, key)
		}
	}
	return dates, nil
}

func (r *WorkoutSessionRepository) Cancel(ctx context.Context, id, userID uint) (*models.WorkoutSession, error) {
	var s models.WorkoutSession
	err := r.DB().WithContext(ctx).Where("id = ? AND user_id = ? AND ended_at IS NULL", id, userID).First(&s).Error
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	err = r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exercises []models.WorkoutSessionExercise
		if err := tx.Preload("Sets").Where("workout_session_id = ?", id).Find(&exercises).Error; err != nil {
			return err
		}

		for _, ex := range exercises {
			var incompleteSetIDs []uint
			hasCompleted := false
			for _, set := range ex.Sets {
				if set.CompletedAt == nil {
					incompleteSetIDs = append(incompleteSetIDs, set.ID)
				} else {
					hasCompleted = true
				}
			}

			if len(incompleteSetIDs) > 0 {
				if err := tx.Where("workout_session_set_id IN ?", incompleteSetIDs).Delete(&models.WorkoutSessionSetValue{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", incompleteSetIDs).Delete(&models.WorkoutSessionSet{}).Error; err != nil {
					return err
				}
			}

			if !hasCompleted {
				if err := tx.Delete(&models.WorkoutSessionExercise{}, ex.ID).Error; err != nil {
					return err
				}
			}
		}

		return tx.Model(&s).Update("ended_at", now).Error
	})
	if err != nil {
		return nil, err
	}
	s.EndedAt = &now
	return r.GetByID(ctx, s.ID, userID)
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

func (r *WorkoutSessionRepository) DeleteByID(ctx context.Context, id, userID uint) error {
	result := r.DB().WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.WorkoutSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
