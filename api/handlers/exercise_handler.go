package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"liift/api/types"
	"liift/internal/models"
	"liift/internal/repository"
	"liift/internal/utils"

	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	repo        *repository.ExerciseRepository
	imageRepo   *repository.ImageRepository
	storagePath string
}

func NewExerciseHandler(repo *repository.ExerciseRepository, imageRepo *repository.ImageRepository, storagePath string) *ExerciseHandler {
	return &ExerciseHandler{
		repo:        repo,
		imageRepo:   imageRepo,
		storagePath: storagePath,
	}
}

type ExerciseDataResponse struct {
	ID                    uint                 `json:"id"`
	Name                  string               `json:"name"`
	Description           string               `json:"description"`
	Image                 string               `json:"image,omitempty"`
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
	Name                  string   `form:"name" json:"name"`
	Description           string   `form:"description" json:"description,omitempty"`
	ImageGUID             *string  `form:"image_guid" json:"image_guid,omitempty"`
	PrimaryMuscleGroups   []string `form:"primary_muscle_groups" json:"primary_muscle_groups"`
	SecondaryMuscleGroups []string `form:"secondary_muscle_groups" json:"secondary_muscle_groups,omitempty"`
	Equipment             []string `form:"equipment" json:"equipment"`
}

type UpdateExerciseRequest struct {
	Name                  string   `form:"name" json:"name"`
	Description           string   `form:"description" json:"description,omitempty"`
	ImageGUID             *string  `form:"image_guid" json:"image_guid,omitempty"`
	PrimaryMuscleGroups   []string `form:"primary_muscle_groups" json:"primary_muscle_groups"`
	SecondaryMuscleGroups []string `form:"secondary_muscle_groups" json:"secondary_muscle_groups,omitempty"`
	Equipment             []string `form:"equipment" json:"equipment"`
}

func (h *ExerciseHandler) handleImageUpload(c echo.Context, formFieldName string) (*string, error) {
	file, err := c.FormFile(formFieldName)
	if err != nil {
		if err.Error() != "http: no such file" {
			c.Logger().Debugf("No file uploaded in field '%s': %v", formFieldName, err)
		}
		return nil, nil
	}

	// Validate file size (e.g., max 10MB)
	if file.Size > 10*1024*1024 {
		c.Logger().Warnf("Image file too large: %d bytes", file.Size)
		return nil, fmt.Errorf("file_too_large")
	}

	// Validate MIME type
	src, err := file.Open()
	if err != nil {
		c.Logger().Errorf("Failed to open uploaded file: %v", err)
		return nil, fmt.Errorf("failed_to_open_file")
	}
	defer func() {
		if err := src.Close(); err != nil {
			c.Logger().Errorf("Error closing source connection", err)
		}
	}()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		c.Logger().Errorf("Failed to read file for MIME detection: %v", err)
		return nil, fmt.Errorf("failed_to_read_file")
	}

	mimeType := http.DetectContentType(buffer)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" && mimeType != "image/webp" {
		c.Logger().Warnf("Invalid image MIME type: %s", mimeType)
		return nil, fmt.Errorf("invalid_image_type")
	}

	_, err = src.Seek(0, 0)
	if err != nil {
		c.Logger().Errorf("Failed to reset file pointer: %v", err)
		return nil, fmt.Errorf("failed_to_process_file")
	}

	guid, err := utils.GenerateGUID()
	if err != nil {
		c.Logger().Errorf("Failed to generate GUID for image: %v", err)
		return nil, fmt.Errorf("failed_to_generate_guid")
	}

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	default:
		ext = ".jpg"
	}

	if err := os.MkdirAll(h.storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed_to_create_storage_directory")
	}
	// Save file with GUID as filename
	filename := guid + ext
	filePath := filepath.Join(h.storagePath, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed_to_save_file")
	}
	defer func() {
		if err := dst.Close(); err != nil {
			c.Logger().Errorf("Error closing destination connection", err)
		}
	}()

	if _, err := io.Copy(dst, src); err != nil {
		if err = os.Remove(filePath); err != nil {
			c.Logger().Errorf("Failed to clean up file after write failure: %v", err)
		}
		return nil, fmt.Errorf("failed_to_write_file")
	}

	image := &models.Image{
		GUID:     guid,
		Path:     filePath,
		MimeType: mimeType,
	}

	if err := h.imageRepo.Create(c.Request().Context(), image); err != nil {
		c.Logger().Errorf("Failed to save image record to database (GUID: %s, Path: %s): %v. Cleaning up file.", guid, filePath, err)
		if err = os.Remove(filePath); err != nil {
			c.Logger().Errorf("Failed to clean up file after save failure: %v", err)
		}
		return nil, fmt.Errorf("failed_to_save_image_record")
	}

	return &guid, nil
}

func (h *ExerciseHandler) deleteImageIfExists(ctx echo.Context, guid *string) error {
	if guid == nil || *guid == "" {
		return nil
	}

	image, err := h.imageRepo.GetByGUID(ctx.Request().Context(), *guid)
	if err != nil {
		return nil
	}

	if err := os.Remove(image.Path); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to delete image file: %v\n", err)
	}

	if err := h.imageRepo.Delete(ctx.Request().Context(), *guid); err != nil {
		fmt.Printf("Warning: failed to delete image record: %v\n", err)
	}

	return nil
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

	search := c.QueryParam("q")
	muscleGroups := c.QueryParams()["muscle_group"]
	equipment := c.QueryParams()["equipment"]

	exercises, total, err := h.repo.List(c.Request().Context(), limit, offset, search, muscleGroups, equipment)
	if err != nil {
		c.Logger().Errorf("Failed to fetch exercises: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "failed_to_fetch_exercises",
		})
	}

	data := utils.Map(exercises, func(exercise models.Exercise) ExerciseDataResponse {
		var imagePath string
		if exercise.ImageGUID != nil && *exercise.ImageGUID != "" {
			imagePath = "/api/images/" + *exercise.ImageGUID
		}
		return ExerciseDataResponse{
			ID:                    exercise.ID,
			Name:                  exercise.Name,
			Description:           exercise.Description,
			Image:                 imagePath,
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
		c.Logger().Errorf("Failed to get exercise by ID %d: %v", id, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "exercises_not_found",
		})
	}

	return c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) CreateExercise(c echo.Context) error {
	var req CreateExerciseRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind create exercise request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	var imageGUID *string
	if imageGUIDFromFile, err := h.handleImageUpload(c, "image"); err != nil {
		if err.Error() != "" {
			c.Logger().Errorf("Failed to handle image upload: %v", err)
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
		}
	} else if imageGUIDFromFile != nil {
		imageGUID = imageGUIDFromFile
	} else {
		imageGUID = req.ImageGUID
	}

	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
		ImageGUID:   imageGUID,
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
		c.Logger().Errorf("Failed to create exercise: %v", err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_creation_failed",
		})
	}

	created, err := h.repo.GetByID(c.Request().Context(), exercise.ID)
	if err != nil {
		c.Logger().Errorf("Failed to fetch created exercise (ID: %d): %v", exercise.ID, err)
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "exercise_fetch_failed",
		})
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *ExerciseHandler) UpdateExercise(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Logger().Errorf("Invalid exercise ID in update request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_exercise_id",
		})
	}

	var req UpdateExerciseRequest

	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("Failed to bind update exercise request: %v", err)
		c.Logger().Errorf("Content-Type: %s", c.Request().Header.Get("Content-Type"))
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_request_body",
		})
	}

	c.Logger().Debugf("Bound request - Name: %s, PrimaryGroups: %v, Equipment: %v",
		req.Name, req.PrimaryMuscleGroups, req.Equipment)

	existingExercise, err := h.repo.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		c.Logger().Errorf("Failed to get exercise by ID %d for update: %v", id, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "exercise_not_found",
		})
	}

	var imageGUID *string
	imageGUIDFromFile, err := h.handleImageUpload(c, "image")
	if err != nil {
		c.Logger().Errorf("Failed to handle image upload during update (exercise ID: %d): %v", id, err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: err.Error(),
		})
	}
	if imageGUIDFromFile != nil {
		if existingExercise.ImageGUID != nil {
			if err = h.deleteImageIfExists(c, existingExercise.ImageGUID); err != nil {
				c.Logger().Warnf("Failed to delete existing image during update (exercise ID: %d): %v", id, err)
			}
		}
		imageGUID = imageGUIDFromFile
	} else {
		if req.ImageGUID != nil && *req.ImageGUID == "" {
			// Clear the image
			if existingExercise.ImageGUID != nil {
				if err = h.deleteImageIfExists(c, existingExercise.ImageGUID); err != nil {
					c.Logger().Warnf("Failed to delete existing image during update (exercise ID: %d): %v", id, err)
				}
			}
			imageGUID = nil
		} else if req.ImageGUID != nil && *req.ImageGUID != "" {
			imageGUID = req.ImageGUID
			if existingExercise.ImageGUID != nil && *existingExercise.ImageGUID != *req.ImageGUID {
				if err = h.deleteImageIfExists(c, existingExercise.ImageGUID); err != nil {
					c.Logger().Warnf("Failed to delete existing image during update (exercise ID: %d): %v", id, err)
				}
			}
		} else {
			imageGUID = existingExercise.ImageGUID
		}
	}

	exercise := models.Exercise{
		Name:        req.Name,
		Description: req.Description,
		ImageGUID:   imageGUID,
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
		c.Logger().Errorf("Failed to update exercise (ID: %d): %v", id, err)
		c.Logger().Errorf("Exercise data: Name=%s, ImageGUID=%v, PrimaryGroups=%d, Equipment=%d",
			exercise.Name, exercise.ImageGUID, len(exercise.PrimaryMuscleGroups), len(exercise.Equipment))
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
		c.Logger().Errorf("Invalid exercise ID in delete request: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_exercise_id",
		})
	}

	if err := h.repo.Delete(c.Request().Context(), uint(id)); err != nil {
		c.Logger().Errorf("Failed to delete exercise (ID: %d): %v", id, err)
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
