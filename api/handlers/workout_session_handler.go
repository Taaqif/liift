package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"liift/api/types"
	"liift/api/middleware"
	"liift/internal/models"
	"liift/internal/repository"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type optionalTime struct {
	t *time.Time
}

func (o *optionalTime) UnmarshalJSON(data []byte) error {
	o.t = nil
	if len(data) == 0 {
		return nil
	}
	if len(data) == 4 && string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
	}
	o.t = &parsed
	return nil
}

func (o *optionalTime) Time() *time.Time {
	return o.t
}

type WorkoutSessionHandler struct {
	repo            *repository.WorkoutSessionRepository
	planProgressRepo *repository.WorkoutPlanProgressRepository
}

func NewWorkoutSessionHandler(repo *repository.WorkoutSessionRepository, planProgressRepo *repository.WorkoutPlanProgressRepository) *WorkoutSessionHandler {
	return &WorkoutSessionHandler{repo: repo, planProgressRepo: planProgressRepo}
}

type WorkoutSessionSetValueResponse struct {
	ID          uint    `json:"id"`
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type WorkoutSessionSetResponse struct {
	ID          uint                            `json:"id"`
	WorkoutSetID *uint                          `json:"workout_set_id,omitempty"`
	Order       int                             `json:"order"`
	CompletedAt *time.Time                      `json:"completed_at"`
	Values      []WorkoutSessionSetValueResponse `json:"values"`
}

type WorkoutSessionExerciseRefResponse struct {
	ID                    uint              `json:"id"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	Instructions          []string          `json:"instructions,omitempty"`
	Image                 string            `json:"image,omitempty"`
	PrimaryMuscleGroups   []ExerciseRefItem `json:"primary_muscle_groups"`
	SecondaryMuscleGroups []ExerciseRefItem `json:"secondary_muscle_groups"`
	Equipment             []ExerciseRefItem `json:"equipment"`
	ExerciseFeatures      []ExerciseRefItem `json:"exercise_features"`
}

type WorkoutSessionExerciseResponse struct {
	ID                 uint                            `json:"id"`
	WorkoutExerciseID  uint                            `json:"workout_exercise_id"`
	Order              int                             `json:"order"`
	Note               string                          `json:"note"`
	RestTimer          int                             `json:"rest_timer"`
	Exercise           *WorkoutSessionExerciseRefResponse `json:"exercise,omitempty"`
	Sets               []WorkoutSessionSetResponse    `json:"sets"`
}

type WorkoutSessionWorkoutRefResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WorkoutSessionResponse struct {
	ID             uint                             `json:"id"`
	UserID         uint                             `json:"user_id"`
	WorkoutID      uint                             `json:"workout_id"`
	PlanProgressID *uint                            `json:"plan_progress_id,omitempty"`
	StartedAt      time.Time                        `json:"started_at"`
	EndedAt        *time.Time                       `json:"ended_at"`
	Workout        *WorkoutSessionWorkoutRefResponse `json:"workout,omitempty"`
	Exercises      []WorkoutSessionExerciseResponse `json:"exercises"`
}

func mapSessionSetValueToResponse(v models.WorkoutSessionSetValue) WorkoutSessionSetValueResponse {
	return WorkoutSessionSetValueResponse{
		ID:          v.ID,
		FeatureName: v.FeatureName,
		Value:       v.Value,
	}
}

func mapSessionSetToResponse(s models.WorkoutSessionSet) WorkoutSessionSetResponse {
	return WorkoutSessionSetResponse{
		ID:          s.ID,
		WorkoutSetID: s.WorkoutSetID,
		Order:       s.Order,
		CompletedAt: s.CompletedAt,
		Values:      utils.Map(s.Values, mapSessionSetValueToResponse),
	}
}

func mapSessionExerciseToRefResponse(ex *models.Exercise) *WorkoutSessionExerciseRefResponse {
	if ex == nil || ex.ID == 0 {
		return nil
	}
	primary, secondary, equipment, features := mapExerciseAssociationsToRefItems(ex)
	return &WorkoutSessionExerciseRefResponse{
		ID:                    ex.ID,
		Name:                  ex.Name,
		Description:           ex.Description,
		Instructions:          ex.Instructions,
		Image:                 buildExerciseImagePath(ex.ImageGUID),
		PrimaryMuscleGroups:   primary,
		SecondaryMuscleGroups: secondary,
		Equipment:             equipment,
		ExerciseFeatures:      features,
	}
}

func mapSessionExerciseToResponse(e models.WorkoutSessionExercise) WorkoutSessionExerciseResponse {
	res := WorkoutSessionExerciseResponse{
		ID:                e.ID,
		WorkoutExerciseID:  e.WorkoutExerciseID,
		Order:             e.Order,
		Note:              e.Note,
		RestTimer:         e.RestTimer,
		Sets:              utils.Map(e.Sets, mapSessionSetToResponse),
	}
	if e.WorkoutExerciseID != 0 && e.WorkoutExercise.ID != 0 {
		if e.WorkoutExercise.Exercise.ID != 0 {
			res.Exercise = mapSessionExerciseToRefResponse(&e.WorkoutExercise.Exercise)
		}
	} else if e.ExerciseID != nil && e.Exercise.ID != 0 {
		res.Exercise = mapSessionExerciseToRefResponse(&e.Exercise)
	}
	return res
}

func mapSessionToResponse(s *models.WorkoutSession) WorkoutSessionResponse {
	res := WorkoutSessionResponse{
		ID:             s.ID,
		UserID:         s.UserID,
		WorkoutID:      s.WorkoutID,
		PlanProgressID: s.PlanProgressID,
		StartedAt:      s.StartedAt,
		EndedAt:        s.EndedAt,
		Exercises:      utils.Map(s.Exercises, mapSessionExerciseToResponse),
	}
	if s.Workout.ID != 0 {
		res.Workout = &WorkoutSessionWorkoutRefResponse{
			ID:          s.Workout.ID,
			Name:        s.Workout.Name,
			Description: s.Workout.Description,
		}
	}
	return res
}

type UpdateWorkoutSessionRequest struct {
	Exercises []UpdateWorkoutSessionExerciseRequest `json:"exercises"`
}

type UpdateWorkoutSessionSetValueRequest struct {
	ID          *uint   `json:"id,omitempty"`
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type UpdateWorkoutSessionSetRequest struct {
	ID          *uint                                 `json:"id,omitempty"`
	WorkoutSetID *uint                                `json:"workout_set_id,omitempty"`
	Order       int                                   `json:"order"`
	CompletedAt optionalTime                          `json:"completed_at"`
	Values      []UpdateWorkoutSessionSetValueRequest `json:"values"`
}

type UpdateWorkoutSessionExerciseRequest struct {
	ID                *uint                            `json:"id,omitempty"`
	WorkoutExerciseID uint                             `json:"workout_exercise_id"`
	Order             int                              `json:"order"`
	Note              string                           `json:"note"`
	RestTimer         int                              `json:"rest_timer"`
	Sets              []UpdateWorkoutSessionSetRequest `json:"sets"`
}

func (h *WorkoutSessionHandler) StartWorkout(c echo.Context) error {
	workoutID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_workout_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	session, err := h.repo.Start(c.Request().Context(), userID, uint(workoutID))
	if err != nil {
		if err == repository.ErrActiveSessionExists {
			return c.JSON(http.StatusConflict, types.ErrorResponse{Error: "active_session_exists"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "workout_not_found"})
		}
		c.Logger().Errorf("Failed to start workout session: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_start_session"})
	}

	return c.JSON(http.StatusCreated, mapSessionToResponse(session))
}

func (h *WorkoutSessionHandler) GetActive(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	session, err := h.repo.GetActiveByUserID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "no_active_session"})
		}
		c.Logger().Errorf("Failed to get active session: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_session"})
	}

	return c.JSON(http.StatusOK, mapSessionToResponse(session))
}

func (h *WorkoutSessionHandler) GetSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_session_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	session, err := h.repo.GetByID(c.Request().Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
		}
		c.Logger().Errorf("Failed to get session %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_session"})
	}

	return c.JSON(http.StatusOK, mapSessionToResponse(session))
}

func (h *WorkoutSessionHandler) UpdateSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_session_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	var req UpdateWorkoutSessionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	session := &models.WorkoutSession{}
	session.ID = uint(id)
	session.UserID = userID
	for _, exReq := range req.Exercises {
		ex := models.WorkoutSessionExercise{
			WorkoutExerciseID: exReq.WorkoutExerciseID,
			Order:             exReq.Order,
			Note:              exReq.Note,
			RestTimer:         exReq.RestTimer,
		}
		if exReq.ID != nil {
			ex.ID = *exReq.ID
		}
		for _, setReq := range exReq.Sets {
			set := models.WorkoutSessionSet{
				WorkoutSetID: setReq.WorkoutSetID,
				Order:       setReq.Order,
				CompletedAt: setReq.CompletedAt.Time(),
			}
			if setReq.ID != nil {
				set.ID = *setReq.ID
			}
			for _, vReq := range setReq.Values {
				v := models.WorkoutSessionSetValue{
					FeatureName: vReq.FeatureName,
					Value:       vReq.Value,
				}
				if vReq.ID != nil {
					v.ID = *vReq.ID
				}
				set.Values = append(set.Values, v)
			}
			ex.Sets = append(ex.Sets, set)
		}
		session.Exercises = append(session.Exercises, ex)
	}

	if err := h.repo.Update(c.Request().Context(), session, userID); err != nil {
		if err.Error() == "cannot update ended session" {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "cannot_update_ended_session"})
		}
		c.Logger().Errorf("Failed to update session %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_update_session"})
	}

	updated, _ := h.repo.GetByID(c.Request().Context(), uint(id), userID)
	return c.JSON(http.StatusOK, mapSessionToResponse(updated))
}

type AddExerciseToSessionRequest struct {
	ExerciseID uint `json:"exercise_id"`
	RestTimer  int  `json:"rest_timer"`
}

func (h *WorkoutSessionHandler) AddExercise(c echo.Context) error {
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_session_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	var req AddExerciseToSessionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}
	if req.ExerciseID == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "exercise_id_required"})
	}

	session, err := h.repo.AddExerciseToSession(c.Request().Context(), uint(sessionID), userID, req.ExerciseID, req.RestTimer)
	if err != nil {
		if err.Error() == "cannot add exercise to ended session" {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "cannot_add_to_ended_session"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_or_exercise_not_found"})
		}
		c.Logger().Errorf("Failed to add exercise to session %d: %v", sessionID, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_add_exercise"})
	}

	return c.JSON(http.StatusCreated, mapSessionToResponse(session))
}

func (h *WorkoutSessionHandler) CancelSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_session_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	session, err := h.repo.Cancel(c.Request().Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
		}
		c.Logger().Errorf("Failed to cancel session %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_cancel_session"})
	}

	return c.JSON(http.StatusOK, mapSessionToResponse(session))
}

func (h *WorkoutSessionHandler) EndSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_session_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	session, err := h.repo.End(c.Request().Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
		}
		c.Logger().Errorf("Failed to end session %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_end_session"})
	}

	if session.PlanProgressID != nil && h.planProgressRepo != nil {
		if err := h.planProgressRepo.AdvanceDay(c.Request().Context(), *session.PlanProgressID); err != nil {
			c.Logger().Errorf("Failed to advance plan progress %d: %v", *session.PlanProgressID, err)
		}
	}

	return c.JSON(http.StatusOK, mapSessionToResponse(session))
}

func RegisterWorkoutSessionRoutes(api *echo.Group, handler *WorkoutSessionHandler) {
	api.POST("/workouts/:id/start", handler.StartWorkout)
	api.GET("/workout-sessions/active", handler.GetActive)
	api.POST("/workout-sessions/:id/exercises", handler.AddExercise)
	api.POST("/workout-sessions/:id/end", handler.EndSession)
	api.POST("/workout-sessions/:id/cancel", handler.CancelSession)
	api.GET("/workout-sessions/:id", handler.GetSession)
	api.PATCH("/workout-sessions/:id", handler.UpdateSession)
}
