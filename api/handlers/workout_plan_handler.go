package handlers

import (
	"net/http"
	"strconv"
	"time"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/repository"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
)

type WorkoutPlanHandler struct {
	repo *repository.WorkoutPlanRepository
}

func NewWorkoutPlanHandler(repo *repository.WorkoutPlanRepository) *WorkoutPlanHandler {
	return &WorkoutPlanHandler{repo: repo}
}

type PlanDayRequest struct {
	WorkoutIDs  []uint `json:"workoutIds"`
	Description string `json:"description"`
}

type PlanWeekRequest struct {
	Days []PlanDayRequest `json:"days"`
}

type CreateWorkoutPlanRequest struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	NumberOfWeeks int               `json:"numberOfWeeks"`
	DaysPerWeek   int               `json:"daysPerWeek"`
	Weeks         []PlanWeekRequest `json:"weeks"`
}

type UpdateWorkoutPlanRequest struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	NumberOfWeeks int               `json:"numberOfWeeks"`
	DaysPerWeek   int               `json:"daysPerWeek"`
	Weeks         []PlanWeekRequest `json:"weeks"`
}

type PlanDayResponse struct {
	WorkoutIDs  []uint `json:"workoutIds"`
	Description string `json:"description,omitempty"`
}

type PlanWeekResponse struct {
	Days []PlanDayResponse `json:"days"`
}

type WorkoutPlanResponse struct {
	ID            uint                `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	NumberOfWeeks int                 `json:"numberOfWeeks"`
	DaysPerWeek   int                 `json:"daysPerWeek"`
	Weeks         []PlanWeekResponse   `json:"weeks"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type WorkoutPlansListResponse struct {
	Data   []WorkoutPlanResponse `json:"data"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func mapScheduleToWeeks(s models.ScheduleData) []PlanWeekResponse {
	out := make([]PlanWeekResponse, len(s))
	for i, w := range s {
		out[i] = PlanWeekResponse{
			Days: utils.Map(w.Days, func(d models.PlanDayJSON) PlanDayResponse {
				ids := make([]uint, len(d.WorkoutIDs))
				copy(ids, d.WorkoutIDs)
				return PlanDayResponse{WorkoutIDs: ids, Description: d.Description}
			}),
		}
	}
	return out
}

func mapWorkoutPlanToResponse(p *models.WorkoutPlan) WorkoutPlanResponse {
	return WorkoutPlanResponse{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		NumberOfWeeks: p.NumberOfWeeks,
		DaysPerWeek:   p.DaysPerWeek,
		Weeks:         mapScheduleToWeeks(p.Schedule),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func requestWeeksToSchedule(weeks []PlanWeekRequest) models.ScheduleData {
	out := make(models.ScheduleData, len(weeks))
	for i, w := range weeks {
		out[i] = models.PlanWeekJSON{
			Days: utils.Map(w.Days, func(d PlanDayRequest) models.PlanDayJSON {
				ids := make([]uint, len(d.WorkoutIDs))
				copy(ids, d.WorkoutIDs)
				return models.PlanDayJSON{WorkoutIDs: ids, Description: d.Description}
			}),
		}
	}
	return out
}

func (h *WorkoutPlanHandler) GetWorkoutPlans(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	plans, total, err := h.repo.List(c.Request().Context(), limit, offset)
	if err != nil {
		c.Logger().Errorf("Failed to list workout plans: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed_to_fetch_workout_plans",
		})
	}

	data := utils.Map(plans, func(p models.WorkoutPlan) WorkoutPlanResponse {
		return mapWorkoutPlanToResponse(&p)
	})
	return c.JSON(http.StatusOK, WorkoutPlansListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *WorkoutPlanHandler) GetWorkoutPlan(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}

	plan, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		c.Logger().Errorf("Failed to get workout plan %d: %v", id, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "workout_plan_not_found",
		})
	}
	return c.JSON(http.StatusOK, mapWorkoutPlanToResponse(plan))
}

func (h *WorkoutPlanHandler) CreateWorkoutPlan(c echo.Context) error {
	var req CreateWorkoutPlanRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind create workout plan request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	plan := models.WorkoutPlan{
		Name:          req.Name,
		Description:   req.Description,
		NumberOfWeeks: req.NumberOfWeeks,
		DaysPerWeek:   req.DaysPerWeek,
		Schedule:      requestWeeksToSchedule(req.Weeks),
	}

	if plan.Name == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "name_required"})
	}
	if plan.NumberOfWeeks < 1 || plan.NumberOfWeeks > 52 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_number_of_weeks"})
	}
	if plan.DaysPerWeek < 1 || plan.DaysPerWeek > 14 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_days_per_week"})
	}
	for _, w := range plan.Schedule {
		for _, d := range w.Days {
			if len(d.WorkoutIDs) == 0 {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "each_day_requires_at_least_one_workout"})
			}
		}
	}

	if err := h.repo.Create(c.Request().Context(), &plan); err != nil {
		c.Logger().Errorf("Failed to create workout plan: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_plan_creation_failed",
		})
	}

	created, err := h.repo.GetByID(c.Request().Context(), plan.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_plan_fetch_failed",
		})
	}
	return c.JSON(http.StatusCreated, mapWorkoutPlanToResponse(created))
}

func (h *WorkoutPlanHandler) UpdateWorkoutPlan(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}

	var req UpdateWorkoutPlanRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind update workout plan request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	_, err = h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "workout_plan_not_found",
		})
	}

	plan := models.WorkoutPlan{
		Name:          req.Name,
		Description:   req.Description,
		NumberOfWeeks: req.NumberOfWeeks,
		DaysPerWeek:   req.DaysPerWeek,
		Schedule:      requestWeeksToSchedule(req.Weeks),
	}
	plan.ID = uint(id)

	if plan.Name == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "name_required"})
	}
	if plan.NumberOfWeeks < 1 || plan.NumberOfWeeks > 52 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_number_of_weeks"})
	}
	if plan.DaysPerWeek < 1 || plan.DaysPerWeek > 14 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_days_per_week"})
	}
	for _, w := range plan.Schedule {
		for _, d := range w.Days {
			if len(d.WorkoutIDs) == 0 {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "each_day_requires_at_least_one_workout"})
			}
		}
	}

	if err := h.repo.Update(c.Request().Context(), &plan); err != nil {
		c.Logger().Errorf("Failed to update workout plan %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_plan_update_failed",
		})
	}

	updated, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_plan_fetch_failed",
		})
	}
	return c.JSON(http.StatusOK, mapWorkoutPlanToResponse(updated))
}

func (h *WorkoutPlanHandler) DeleteWorkoutPlan(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_id"})
	}

	if err := h.repo.Delete(c.Request().Context(), uint(id)); err != nil {
		c.Logger().Errorf("Failed to delete workout plan %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_plan_deletion_failed",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func RegisterWorkoutPlanRoutes(api *echo.Group, handler *WorkoutPlanHandler) {
	api.GET("/workout-plans", handler.GetWorkoutPlans)
	api.GET("/workout-plans/:id", handler.GetWorkoutPlan)
	api.POST("/workout-plans", handler.CreateWorkoutPlan)
	api.PUT("/workout-plans/:id", handler.UpdateWorkoutPlan)
	api.DELETE("/workout-plans/:id", handler.DeleteWorkoutPlan)
}
