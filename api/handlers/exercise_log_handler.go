package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"liift/api/middleware"
	"liift/api/types"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ExerciseLogHandler struct {
	exerciseRepo *repository.ExerciseRepository
	sessionRepo  *repository.WorkoutSessionRepository
}

func NewExerciseLogHandler(exerciseRepo *repository.ExerciseRepository, sessionRepo *repository.WorkoutSessionRepository) *ExerciseLogHandler {
	return &ExerciseLogHandler{exerciseRepo: exerciseRepo, sessionRepo: sessionRepo}
}

type ExerciseLogValueResponse struct {
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type ExerciseLogSetResponse struct {
	Order       int                        `json:"order"`
	CompletedAt string                     `json:"completed_at"`
	Values      []ExerciseLogValueResponse `json:"values"`
}

type ExerciseLogEntryResponse struct {
	SessionID   uint                     `json:"session_id"`
	Date        string                   `json:"date"`
	WorkoutName string                   `json:"workout_name"`
	Sets        []ExerciseLogSetResponse `json:"sets"`
}

type ExerciseLogsResponse struct {
	Data   []ExerciseLogEntryResponse `json:"data"`
	Total  int64                      `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}

func (h *ExerciseLogHandler) GetLogs(c echo.Context) error {
	exerciseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_exercise_id"})
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: "authorization_header_missing"})
	}

	if _, err := h.exerciseRepo.GetByID(c.Request().Context(), uint(exerciseID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "exercise_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_exercise"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	entries, total, err := h.sessionRepo.GetExerciseLogs(c.Request().Context(), uint(exerciseID), userID, limit, offset)
	if err != nil {
		c.Logger().Errorf("Failed to get exercise logs for exercise %d: %v", exerciseID, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_logs"})
	}

	data := make([]ExerciseLogEntryResponse, len(entries))
	for i, e := range entries {
		sets := make([]ExerciseLogSetResponse, len(e.Sets))
		for j, s := range e.Sets {
			vals := make([]ExerciseLogValueResponse, len(s.Values))
			for k, v := range s.Values {
				vals[k] = ExerciseLogValueResponse{FeatureName: v.FeatureName, Value: v.Value}
			}
			sets[j] = ExerciseLogSetResponse{
				Order:       s.Order,
				CompletedAt: s.CompletedAt.UTC().Format("2006-01-02T15:04:05Z"),
				Values:      vals,
			}
		}
		data[i] = ExerciseLogEntryResponse{
			SessionID:   e.SessionID,
			Date:        e.Date.UTC().Format("2006-01-02T15:04:05Z"),
			WorkoutName: e.WorkoutName,
			Sets:        sets,
		}
	}

	return c.JSON(http.StatusOK, ExerciseLogsResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func RegisterExerciseLogRoutes(api *echo.Group, handler *ExerciseLogHandler) {
	api.GET("/exercises/:id/logs", handler.GetLogs)
}
