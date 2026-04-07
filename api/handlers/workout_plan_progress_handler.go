package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"liift/api/middleware"
	"liift/api/types"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Ensure mapSessionToResponse is accessible (defined in workout_session_handler.go)

type WorkoutPlanProgressHandler struct {
	repo        *repository.WorkoutPlanProgressRepository
	planRepo    *repository.WorkoutPlanRepository
	sessionRepo *repository.WorkoutSessionRepository
}

func NewWorkoutPlanProgressHandler(
	repo *repository.WorkoutPlanProgressRepository,
	planRepo *repository.WorkoutPlanRepository,
	sessionRepo *repository.WorkoutSessionRepository,
) *WorkoutPlanProgressHandler {
	return &WorkoutPlanProgressHandler{repo: repo, planRepo: planRepo, sessionRepo: sessionRepo}
}

type WorkoutPlanProgressResponse struct {
	ID          uint                `json:"id"`
	UserID      uint                `json:"user_id"`
	PlanID      uint                `json:"plan_id"`
	Plan        WorkoutPlanResponse `json:"plan"`
	CurrentWeek int                 `json:"current_week"`
	CurrentDay  int                 `json:"current_day"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at"`
}

func (h *WorkoutPlanProgressHandler) GetActive(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	progress, err := h.repo.GetActiveByUserID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "no_active_plan_progress"})
		}
		c.Logger().Errorf("Failed to get active plan progress: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_plan_progress"})
	}

	return c.JSON(http.StatusOK, WorkoutPlanProgressResponse{
		ID:          progress.ID,
		UserID:      progress.UserID,
		PlanID:      progress.PlanID,
		Plan:        mapWorkoutPlanToResponse(&progress.Plan),
		CurrentWeek: progress.CurrentWeek,
		CurrentDay:  progress.CurrentDay,
		StartedAt:   progress.StartedAt,
		CompletedAt: progress.CompletedAt,
	})
}

type StartPlanRequest struct {
	PlanID uint `json:"plan_id"`
}

func (h *WorkoutPlanProgressHandler) Start(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	var req StartPlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}
	if req.PlanID == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "plan_id_required"})
	}

	// Verify plan exists
	if _, err := h.planRepo.GetByID(c.Request().Context(), req.PlanID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "workout_plan_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_plan"})
	}

	progress, err := h.repo.Start(c.Request().Context(), userID, req.PlanID)
	if err != nil {
		if errors.Is(err, repository.ErrActivePlanProgressExists) {
			return c.JSON(http.StatusConflict, types.ErrorResponse{Error: "active_plan_progress_exists"})
		}
		c.Logger().Errorf("Failed to start plan progress: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_start_plan"})
	}

	return c.JSON(http.StatusCreated, WorkoutPlanProgressResponse{
		ID:          progress.ID,
		UserID:      progress.UserID,
		PlanID:      progress.PlanID,
		Plan:        mapWorkoutPlanToResponse(&progress.Plan),
		CurrentWeek: progress.CurrentWeek,
		CurrentDay:  progress.CurrentDay,
		StartedAt:   progress.StartedAt,
		CompletedAt: progress.CompletedAt,
	})
}

type UpdatePositionRequest struct {
	CurrentWeek int `json:"current_week"`
	CurrentDay  int `json:"current_day"`
}

func (h *WorkoutPlanProgressHandler) UpdatePosition(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	var req UpdatePositionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	progress, err := h.repo.UpdatePosition(c.Request().Context(), uint(id), userID, req.CurrentWeek, req.CurrentDay)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "plan_progress_not_found"})
		}
		c.Logger().Errorf("Failed to update plan progress position: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_update_position"})
	}

	return c.JSON(http.StatusOK, WorkoutPlanProgressResponse{
		ID:          progress.ID,
		UserID:      progress.UserID,
		PlanID:      progress.PlanID,
		Plan:        mapWorkoutPlanToResponse(&progress.Plan),
		CurrentWeek: progress.CurrentWeek,
		CurrentDay:  progress.CurrentDay,
		StartedAt:   progress.StartedAt,
		CompletedAt: progress.CompletedAt,
	})
}

func (h *WorkoutPlanProgressHandler) Complete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	progress, err := h.repo.Complete(c.Request().Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "plan_progress_not_found"})
		}
		c.Logger().Errorf("Failed to complete plan progress: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_complete_plan"})
	}

	return c.JSON(http.StatusOK, WorkoutPlanProgressResponse{
		ID:          progress.ID,
		UserID:      progress.UserID,
		PlanID:      progress.PlanID,
		Plan:        mapWorkoutPlanToResponse(&progress.Plan),
		CurrentWeek: progress.CurrentWeek,
		CurrentDay:  progress.CurrentDay,
		StartedAt:   progress.StartedAt,
		CompletedAt: progress.CompletedAt,
	})
}

func (h *WorkoutPlanProgressHandler) Stop(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	if err := h.repo.Stop(c.Request().Context(), uint(id), userID); err != nil {
		c.Logger().Errorf("Failed to stop plan progress: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_stop_plan"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *WorkoutPlanProgressHandler) StartDay(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	progress, err := h.repo.GetByID(c.Request().Context(), uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "plan_progress_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_plan_progress"})
	}
	if progress.CompletedAt != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "plan_already_completed"})
	}

	schedule := progress.Plan.Schedule
	if progress.CurrentWeek >= len(schedule) || progress.CurrentDay >= len(schedule[progress.CurrentWeek].Days) {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_plan_position"})
	}

	day := schedule[progress.CurrentWeek].Days[progress.CurrentDay]
	if len(day.WorkoutIDs) == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "no_workouts_for_day"})
	}

	workoutIDs := make([]uint, len(day.WorkoutIDs))
	copy(workoutIDs, day.WorkoutIDs)

	session, err := h.sessionRepo.StartDay(c.Request().Context(), userID, progress.ID, workoutIDs)
	if err != nil {
		if errors.Is(err, repository.ErrActiveSessionExists) {
			return c.JSON(http.StatusConflict, types.ErrorResponse{Error: "active_session_exists"})
		}
		c.Logger().Errorf("Failed to start plan day session: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_start_session"})
	}

	return c.JSON(http.StatusCreated, mapSessionToResponse(session))
}

func RegisterWorkoutPlanProgressRoutes(api *echo.Group, handler *WorkoutPlanProgressHandler) {
	api.GET("/workout-plan-progress/active", handler.GetActive)
	api.POST("/workout-plan-progress", handler.Start)
	api.PATCH("/workout-plan-progress/:id", handler.UpdatePosition)
	api.POST("/workout-plan-progress/:id/complete", handler.Complete)
	api.POST("/workout-plan-progress/:id/start-day", handler.StartDay)
	api.DELETE("/workout-plan-progress/:id", handler.Stop)
}
