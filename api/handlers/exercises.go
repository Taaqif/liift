package handlers

import (
	"net/http"
	"strconv"

	"liift/api/types"
	"liift/internal/database"
	"liift/internal/models"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

var exerciseRepo *repository.ExerciseRepository

func init() {
	exerciseRepo = repository.NewExerciseRepository(database.DB)
}

type ExercisesListResponse struct {
	Data   []models.Exercise `json:"data"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func GetExercises(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	exercises, total, err := exerciseRepo.List(c.Request().Context(), limit, offset)
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

func GetExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid exercise ID",
		})
	}

	exercise, err := exerciseRepo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "Exercise not found",
		})
	}

	return c.JSON(http.StatusOK, exercise)
}

func CreateExercise(c echo.Context) error {
	var exercise models.Exercise
	if err := c.Bind(&exercise); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := exerciseRepo.Create(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to create exercise",
		})
	}

	return c.JSON(http.StatusCreated, exercise)
}

func UpdateExercise(c echo.Context) error {
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

	if err := exerciseRepo.Update(c.Request().Context(), &exercise); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to update exercise",
		})
	}

	updated, err := exerciseRepo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to fetch updated exercise",
		})
	}

	return c.JSON(http.StatusOK, updated)
}

func DeleteExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid exercise ID",
		})
	}

	if err := exerciseRepo.Delete(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to delete exercise",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func SearchExercises(c echo.Context) error {
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

	exercises, total, err := exerciseRepo.SearchByName(c.Request().Context(), query, limit, offset)
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

func RegisterExerciseRoutes(api *echo.Group) {
	api.GET("/exercises", GetExercises)
	api.GET("/exercises/search", SearchExercises)
	api.GET("/exercises/:id", GetExercise)
	api.POST("/exercises", CreateExercise)
	api.PUT("/exercises/:id", UpdateExercise)
	api.DELETE("/exercises/:id", DeleteExercise)
}
