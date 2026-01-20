package handlers

import (
	"net/http"
	"strconv"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	repo *repository.ExerciseRepository
}

func NewExerciseHandler(repo *repository.ExerciseRepository) *ExerciseHandler {
	return &ExerciseHandler{repo: repo}
}

type ExercisesListResponse struct {
	Data   []models.Exercise `json:"data"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type CreateExerciseRequest struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description,omitempty"`
	PrimaryMuscleGroups   []string `json:"primary_muscle_groups"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups,omitempty"`
	Equipment             []string `json:"equipment"`
}

type UpdateExerciseRequest struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description,omitempty"`
	PrimaryMuscleGroups   []string `json:"primary_muscle_groups"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups,omitempty"`
	Equipment             []string `json:"equipment"`
}

func (h *ExerciseHandler) GetExercises(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	exercises, total, err := h.repo.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed_to_fetch_exercises",
		})
	}

	return c.JSON(http.StatusOK, ExercisesListResponse{
		Data:   exercises,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *ExerciseHandler) GetExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_id",
		})
	}

	exercise, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "exercises_not_found",
		})
	}

	return c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) CreateExercise(c echo.Context) error {
	var req CreateExerciseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "name_required",
		})
	}

	if len(req.PrimaryMuscleGroups) == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "primary_muscle_group_required",
		})
	}

	if len(req.Equipment) == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "equipment_required",
		})
	}

	// Convert request to Exercise model
	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
	}

	// Convert muscle group names to MuscleGroup models
	primaryMuscleGroups := make([]models.MuscleGroup, len(req.PrimaryMuscleGroups))
	for i, name := range req.PrimaryMuscleGroups {
		primaryMuscleGroups[i] = models.MuscleGroup{Name: name}
	}
	exercise.PrimaryMuscleGroups = primaryMuscleGroups

	if len(req.SecondaryMuscleGroups) > 0 {
		secondaryMuscleGroups := make([]models.MuscleGroup, len(req.SecondaryMuscleGroups))
		for i, name := range req.SecondaryMuscleGroups {
			secondaryMuscleGroups[i] = models.MuscleGroup{Name: name}
		}
		exercise.SecondaryMuscleGroups = secondaryMuscleGroups
	}

	// Convert equipment names to Equipment models
	equipment := make([]models.Equipment, len(req.Equipment))
	for i, name := range req.Equipment {
		equipment[i] = models.Equipment{Name: name}
	}
	exercise.Equipment = equipment

	if err := h.repo.Create(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_creation_failed",
		})
	}

	// Reload the exercise with associations to return complete data
	created, err := h.repo.GetByID(c.Request().Context(), exercise.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_fetch_failed",
		})
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *ExerciseHandler) UpdateExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_exercise_id",
		})
	}

	var req UpdateExerciseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "name_required",
		})
	}

	if len(req.PrimaryMuscleGroups) == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "primary_muscle_group_required",
		})
	}

	if len(req.Equipment) == 0 {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "equipment_required",
		})
	}

	// Convert request to Exercise model
	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
	}
	exercise.ID = uint(id)

	// Convert muscle group names to MuscleGroup models
	primaryMuscleGroups := make([]models.MuscleGroup, len(req.PrimaryMuscleGroups))
	for i, name := range req.PrimaryMuscleGroups {
		primaryMuscleGroups[i] = models.MuscleGroup{Name: name}
	}
	exercise.PrimaryMuscleGroups = primaryMuscleGroups

	if len(req.SecondaryMuscleGroups) > 0 {
		secondaryMuscleGroups := make([]models.MuscleGroup, len(req.SecondaryMuscleGroups))
		for i, name := range req.SecondaryMuscleGroups {
			secondaryMuscleGroups[i] = models.MuscleGroup{Name: name}
		}
		exercise.SecondaryMuscleGroups = secondaryMuscleGroups
	}

	// Convert equipment names to Equipment models
	equipment := make([]models.Equipment, len(req.Equipment))
	for i, name := range req.Equipment {
		equipment[i] = models.Equipment{Name: name}
	}
	exercise.Equipment = equipment

	if err := h.repo.Update(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_update_failed",
		})
	}

	updated, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_fetch_failed",
		})
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *ExerciseHandler) DeleteExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_exercise_id",
		})
	}

	if err := h.repo.Delete(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_deletion_failed",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ExerciseHandler) SearchExercises(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "search_query_required",
		})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	exercises, total, err := h.repo.SearchByName(c.Request().Context(), query, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "search_failed",
		})
	}

	return c.JSON(http.StatusOK, ExercisesListResponse{
		Data:   exercises,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func RegisterExerciseRoutes(api *echo.Group, handler *ExerciseHandler) {
	api.GET("/exercises", handler.GetExercises)
	api.GET("/exercises/search", handler.SearchExercises)
	api.GET("/exercises/:id", handler.GetExercise)
	api.POST("/exercises", handler.CreateExercise)
	api.PUT("/exercises/:id", handler.UpdateExercise)
	api.DELETE("/exercises/:id", handler.DeleteExercise)
}
