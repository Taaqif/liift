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

type WorkoutHandler struct {
	repo *repository.WorkoutRepository
}

func NewWorkoutHandler(repo *repository.WorkoutRepository) *WorkoutHandler {
	return &WorkoutHandler{
		repo: repo,
	}
}

type WorkoutSetFeatureRequest struct {
	ID          *uint   `json:"id,omitempty"`
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type WorkoutSetRequest struct {
	ID       *uint                      `json:"id,omitempty"`
	Order    int                        `json:"order"`
	Features []WorkoutSetFeatureRequest `json:"features"`
}

type WorkoutExerciseRequest struct {
	ID         *uint               `json:"id,omitempty"`
	ExerciseID uint                `json:"exercise_id"`
	RestTimer  int                 `json:"rest_timer"`
	Note       string              `json:"note"`
	Order      int                 `json:"order"`
	Sets       []WorkoutSetRequest `json:"sets"`
}

type CreateWorkoutRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	IsLibrary   *bool                    `json:"is_library"`
	Exercises   []WorkoutExerciseRequest `json:"exercises"`
}

type UpdateWorkoutRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Exercises   []WorkoutExerciseRequest `json:"exercises"`
}

type WorkoutSetFeatureResponse struct {
	ID          uint    `json:"id"`
	FeatureName string  `json:"feature_name"`
	Value       float64 `json:"value"`
}

type WorkoutSetResponse struct {
	ID       uint                        `json:"id"`
	Order    int                         `json:"order"`
	Features []WorkoutSetFeatureResponse `json:"features"`
}

type ExerciseRefResponse struct {
	ID                    uint              `json:"id"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	Image                 string            `json:"image,omitempty"`
	PrimaryMuscleGroups   []ExerciseRefItem `json:"primary_muscle_groups"`
	SecondaryMuscleGroups []ExerciseRefItem `json:"secondary_muscle_groups"`
	Equipment             []ExerciseRefItem `json:"equipment"`
	ExerciseFeatures      []ExerciseRefItem `json:"exercise_features"`
}

type WorkoutExerciseResponse struct {
	ID         uint                 `json:"id"`
	WorkoutID  uint                 `json:"workout_id"`
	ExerciseID uint                 `json:"exercise_id"`
	RestTimer  int                  `json:"rest_timer"`
	Note       string               `json:"note"`
	Order      int                  `json:"order"`
	Exercise   *ExerciseRefResponse `json:"exercise,omitempty"`
	Sets       []WorkoutSetResponse `json:"sets"`
}

type WorkoutResponse struct {
	ID          uint                      `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	IsLibrary   bool                      `json:"is_library"`
	Exercises   []WorkoutExerciseResponse `json:"exercises"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type WorkoutsListResponse struct {
	Data   []WorkoutResponse `json:"data"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func mapWorkoutToResponse(w *models.Workout) WorkoutResponse {
	return WorkoutResponse{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		IsLibrary:   w.IsLibrary,
		Exercises:   utils.Map(w.Exercises, mapWorkoutExerciseToResponse),
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func mapWorkoutExerciseToResponse(e models.WorkoutExercise) WorkoutExerciseResponse {
	res := WorkoutExerciseResponse{
		ID:         e.ID,
		WorkoutID:  e.WorkoutID,
		ExerciseID: e.ExerciseID,
		RestTimer:  e.RestTimer,
		Note:       e.Note,
		Order:      e.Order,
		Sets:       utils.Map(e.Sets, mapWorkoutSetToResponse),
	}
	if e.Exercise.ID != 0 {
		res.Exercise = mapExerciseToRefResponse(&e.Exercise)
	}
	return res
}

func mapExerciseToRefResponse(ex *models.Exercise) *ExerciseRefResponse {
	primary, secondary, equipment, features := mapExerciseAssociationsToRefItems(ex)
	return &ExerciseRefResponse{
		ID:                    ex.ID,
		Name:                  ex.Name,
		Description:           ex.Description,
		Image:                 buildExerciseImagePath(ex.ImageGUID),
		PrimaryMuscleGroups:   primary,
		SecondaryMuscleGroups: secondary,
		Equipment:             equipment,
		ExerciseFeatures:      features,
	}
}

func mapWorkoutSetToResponse(s models.WorkoutSet) WorkoutSetResponse {
	return WorkoutSetResponse{
		ID:       s.ID,
		Order:    s.Order,
		Features: utils.Map(s.Features, mapWorkoutSetFeatureToResponse),
	}
}

func mapWorkoutSetFeatureToResponse(f models.WorkoutSetFeature) WorkoutSetFeatureResponse {
	return WorkoutSetFeatureResponse{
		ID:          f.ID,
		FeatureName: f.FeatureName,
		Value:       f.Value,
	}
}

func (h *WorkoutHandler) GetWorkouts(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	search := c.QueryParam("q")
	exerciseFeatures := c.QueryParams()["exercise_feature"]
	exerciseIDStrs := c.QueryParams()["exercise_id"]
	var exerciseIDs []uint
	for _, s := range exerciseIDStrs {
		if id, err := strconv.ParseUint(s, 10, 32); err == nil {
			exerciseIDs = append(exerciseIDs, uint(id))
		}
	}
	muscleGroups := c.QueryParams()["muscle_group"]
	equipment := c.QueryParams()["equipment"]
	includeAll := c.QueryParam("all") == "1"

	workouts, total, err := h.repo.List(c.Request().Context(), limit, offset, search, exerciseFeatures, exerciseIDs, muscleGroups, equipment, includeAll)
	if err != nil {
		c.Logger().Errorf("Failed to fetch workouts: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed_to_fetch_workouts",
		})
	}

	data := utils.Map(workouts, func(w models.Workout) WorkoutResponse {
		return mapWorkoutToResponse(&w)
	})
	return c.JSON(http.StatusOK, WorkoutsListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *WorkoutHandler) GetWorkout(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_id",
		})
	}

	workout, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		c.Logger().Errorf("Failed to get workout by ID %d: %v", id, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "workout_not_found",
		})
	}

	return c.JSON(http.StatusOK, mapWorkoutToResponse(workout))
}

func (h *WorkoutHandler) CreateWorkout(c echo.Context) error {
	var req CreateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind create workout request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	isLibrary := true
	if req.IsLibrary != nil {
		isLibrary = *req.IsLibrary
	}
	workout := models.Workout{
		Name:        req.Name,
		Description: req.Description,
		IsLibrary:   isLibrary,
	}

	exercises := make([]models.WorkoutExercise, len(req.Exercises))
	for i, exReq := range req.Exercises {
		exercise := models.WorkoutExercise{
			ExerciseID: exReq.ExerciseID,
			RestTimer:  exReq.RestTimer,
			Note:       exReq.Note,
			Order:      exReq.Order,
		}

		sets := make([]models.WorkoutSet, len(exReq.Sets))
		for j, setReq := range exReq.Sets {
			set := models.WorkoutSet{
				Order: setReq.Order,
			}

			features := make([]models.WorkoutSetFeature, len(setReq.Features))
			for k, featReq := range setReq.Features {
				features[k] = models.WorkoutSetFeature{
					FeatureName: featReq.FeatureName,
					Value:       featReq.Value,
				}
			}
			set.Features = features
			sets[j] = set
		}
		exercise.Sets = sets
		exercises[i] = exercise
	}
	workout.Exercises = exercises

	if err := workout.Validate(); err != nil {
		c.Logger().Errorf("Workout validation failed (create): %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := h.repo.Create(c.Request().Context(), &workout); err != nil {
		c.Logger().Errorf("Failed to create workout: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_creation_failed",
		})
	}

	created, err := h.repo.GetByID(c.Request().Context(), workout.ID)
	if err != nil {
		c.Logger().Errorf("Failed to fetch created workout (ID: %d): %v", workout.ID, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_fetch_failed",
		})
	}

	return c.JSON(http.StatusCreated, mapWorkoutToResponse(created))
}

func (h *WorkoutHandler) UpdateWorkout(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Logger().Errorf("Invalid workout ID in update request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_workout_id",
		})
	}

	var req UpdateWorkoutRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind update workout request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	_, err = h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		c.Logger().Errorf("Failed to get workout by ID %d for update: %v", id, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "workout_not_found",
		})
	}

	workout := models.Workout{
		Name:        req.Name,
		Description: req.Description,
	}
	workout.ID = uint(id)

	exercises := make([]models.WorkoutExercise, len(req.Exercises))
	for i, exReq := range req.Exercises {
		exercise := models.WorkoutExercise{
			ExerciseID: exReq.ExerciseID,
			RestTimer:  exReq.RestTimer,
			Note:       exReq.Note,
			Order:      exReq.Order,
		}
		if exReq.ID != nil {
			exercise.ID = *exReq.ID
		}

		// Convert sets
		sets := make([]models.WorkoutSet, len(exReq.Sets))
		for j, setReq := range exReq.Sets {
			set := models.WorkoutSet{
				Order: setReq.Order,
			}
			if setReq.ID != nil {
				set.ID = *setReq.ID
			}

			// Convert features
			features := make([]models.WorkoutSetFeature, len(setReq.Features))
			for k, featReq := range setReq.Features {
				features[k] = models.WorkoutSetFeature{
					FeatureName: featReq.FeatureName,
					Value:       featReq.Value,
				}
				if featReq.ID != nil {
					features[k].ID = *featReq.ID
				}
			}
			set.Features = features
			sets[j] = set
		}
		exercise.Sets = sets
		exercises[i] = exercise
	}
	workout.Exercises = exercises

	if err := workout.Validate(); err != nil {
		c.Logger().Errorf("Workout validation failed (update): %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := h.repo.Update(c.Request().Context(), &workout); err != nil {
		c.Logger().Errorf("Failed to update workout (ID: %d): %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_update_failed",
		})
	}

	updated, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_fetch_failed",
		})
	}

	return c.JSON(http.StatusOK, mapWorkoutToResponse(updated))
}

func (h *WorkoutHandler) DeleteWorkout(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Logger().Errorf("Invalid workout ID in delete request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_workout_id",
		})
	}

	if err := h.repo.Delete(c.Request().Context(), uint(id)); err != nil {
		c.Logger().Errorf("Failed to delete workout (ID: %d): %v", id, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "workout_deletion_failed",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func RegisterWorkoutRoutes(api *echo.Group, handler *WorkoutHandler) {
	api.GET("/workouts", handler.GetWorkouts)
	api.GET("/workouts/:id", handler.GetWorkout)
	api.POST("/workouts", handler.CreateWorkout)
	api.PUT("/workouts/:id", handler.UpdateWorkout)
	api.DELETE("/workouts/:id", handler.DeleteWorkout)
}
