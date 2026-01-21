package handlers

import (
	"net/http"
	"strconv"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/repository"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	repo *repository.ExerciseRepository
}

func NewExerciseHandler(repo *repository.ExerciseRepository) *ExerciseHandler {
	return &ExerciseHandler{repo: repo}
}

type ExerciseDataResponse struct {
	ID                    uint                 `json:"id"`
	Name                  string               `json:"name"`
	Description           string               `json:"description"`
	PrimaryMuscleGroups   []models.MuscleGroup `json:"primary_muscle_groups"`
	SecondaryMuscleGroups []models.MuscleGroup `json:"secondary_muscle_groups"`
	Equipment             []models.Equipment   `json:"equipment"`
}

type ExercisesListResponse struct {
	Data   []ExerciseDataResponse `json:"data"`
	Total  int64                  `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
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

	// Optional filters
	search := c.QueryParam("q")
	muscleGroups := c.QueryParams()["muscle_group"]
	equipment := c.QueryParams()["equipment"]

	exercises, total, err := h.repo.List(c.Request().Context(), limit, offset, search, muscleGroups, equipment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed_to_fetch_exercises",
		})
	}

	data := utils.Map(exercises, func(exercise models.Exercise) ExerciseDataResponse {
		return ExerciseDataResponse{
			ID:                    exercise.ID,
			Name:                  exercise.Name,
			Description:           exercise.Description,
			PrimaryMuscleGroups:   exercise.PrimaryMuscleGroups,
			SecondaryMuscleGroups: exercise.SecondaryMuscleGroups,
			Equipment:             exercise.Equipment,
		}
	})

	return c.JSON(http.StatusOK, ExercisesListResponse{
		Data:   data,
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

	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
	}

	primaryMuscleGroups := make([]models.MuscleGroup, len(req.PrimaryMuscleGroups))
	for i, name := range req.PrimaryMuscleGroups {
		mg := models.MuscleGroup{Name: name}
		if err := mg.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
		}
		primaryMuscleGroups[i] = mg
	}
	exercise.PrimaryMuscleGroups = primaryMuscleGroups

	if len(req.SecondaryMuscleGroups) > 0 {
		secondaryMuscleGroups := make([]models.MuscleGroup, len(req.SecondaryMuscleGroups))
		for i, name := range req.SecondaryMuscleGroups {
			mg := models.MuscleGroup{Name: name}
			if err := mg.Validate(); err != nil {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{
					Error: err.Error(),
				})
			}
			secondaryMuscleGroups[i] = mg
		}
		exercise.SecondaryMuscleGroups = secondaryMuscleGroups
	}

	equipment := make([]models.Equipment, len(req.Equipment))
	for i, name := range req.Equipment {
		eq := models.Equipment{Name: name}
		if err := eq.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
		}
		equipment[i] = eq
	}
	exercise.Equipment = equipment

	if err := exercise.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := h.repo.Create(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_creation_failed",
		})
	}

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

	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
	}
	exercise.ID = uint(id)

	primaryMuscleGroups := make([]models.MuscleGroup, len(req.PrimaryMuscleGroups))
	for i, name := range req.PrimaryMuscleGroups {
		mg := models.MuscleGroup{Name: name}
		if err := mg.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
		}
		primaryMuscleGroups[i] = mg
	}
	exercise.PrimaryMuscleGroups = primaryMuscleGroups

	if len(req.SecondaryMuscleGroups) > 0 {
		secondaryMuscleGroups := make([]models.MuscleGroup, len(req.SecondaryMuscleGroups))
		for i, name := range req.SecondaryMuscleGroups {
			mg := models.MuscleGroup{Name: name}
			if err := mg.Validate(); err != nil {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{
					Error: err.Error(),
				})
			}
			secondaryMuscleGroups[i] = mg
		}
		exercise.SecondaryMuscleGroups = secondaryMuscleGroups
	}

	equipment := make([]models.Equipment, len(req.Equipment))
	for i, name := range req.Equipment {
		eq := models.Equipment{Name: name}
		if err := eq.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
		}
		equipment[i] = eq
	}
	exercise.Equipment = equipment

	if err := exercise.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: err.Error(),
		})
	}

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

func RegisterExerciseRoutes(api *echo.Group, handler *ExerciseHandler) {
	api.GET("/exercises", handler.GetExercises)
	api.GET("/exercises/:id", handler.GetExercise)
	api.POST("/exercises", handler.CreateExercise)
	api.PUT("/exercises/:id", handler.UpdateExercise)
	api.DELETE("/exercises/:id", handler.DeleteExercise)
}
