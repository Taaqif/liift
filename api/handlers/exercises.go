package handlers

import (
	"net/http"
	"strconv"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

// ExerciseHandler handles exercise-related HTTP requests
type ExerciseHandler struct {
	repo *repository.ExerciseRepository
}

// NewExerciseHandler creates a new ExerciseHandler with the given repository
func NewExerciseHandler(repo *repository.ExerciseRepository) *ExerciseHandler {
	return &ExerciseHandler{repo: repo}
}

type ExercisesListResponse struct {
	Data   []models.Exercise `json:"data"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
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
			Error: "Failed to fetch exercises",
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
			Error: "Invalid exercise ID",
		})
	}

	exercise, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "Exercise not found",
		})
	}

	return c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) CreateExercise(c echo.Context) error {
	var exercise models.Exercise
	if err := c.Bind(&exercise); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.repo.Create(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to create exercise",
		})
	}

	return c.JSON(http.StatusCreated, exercise)
}

func (h *ExerciseHandler) UpdateExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid exercise ID",
		})
	}

	var exercise models.Exercise
	if err := c.Bind(&exercise); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	exercise.ID = uint(id)

	if err := h.repo.Update(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to update exercise",
		})
	}

	updated, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to fetch updated exercise",
		})
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *ExerciseHandler) DeleteExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid exercise ID",
		})
	}

	if err := h.repo.Delete(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to delete exercise",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ExerciseHandler) SearchExercises(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Search query is required",
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
			Error: "Failed to search exercises",
		})
	}

	return c.JSON(http.StatusOK, ExercisesListResponse{
		Data:   exercises,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// RegisterExerciseRoutes registers exercise routes with the given handler
func RegisterExerciseRoutes(api *echo.Group, handler *ExerciseHandler) {
	api.GET("/exercises", handler.GetExercises)
	api.GET("/exercises/search", handler.SearchExercises)
	api.GET("/exercises/:id", handler.GetExercise)
	api.POST("/exercises", handler.CreateExercise)
	api.PUT("/exercises/:id", handler.UpdateExercise)
	api.DELETE("/exercises/:id", handler.DeleteExercise)
}
