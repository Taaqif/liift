package handlers

import (
	"net/http"
	"os"

	"liift/api/types"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

type ImageHandler struct {
	repo          *repository.ImageRepository
	storagePath   string
	imageBasePath string
}

func NewImageHandler(repo *repository.ImageRepository, storagePath string) *ImageHandler {
	return &ImageHandler{
		repo:          repo,
		storagePath:   storagePath,
		imageBasePath: "/api/images",
	}
}

func (h *ImageHandler) GetImage(c echo.Context) error {
	guid := c.Param("guid")
	if guid == "" {
		c.Logger().Warn("Empty image GUID provided")
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "invalid_image_guid",
		})
	}

	image, err := h.repo.GetByGUID(c.Request().Context(), guid)
	if err != nil {
		c.Logger().Errorf("Failed to get image by GUID %s: %v", guid, err)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "image_not_found",
		})
	}

	if _, err := os.Stat(image.Path); os.IsNotExist(err) {
		c.Logger().Errorf("Image file not found on filesystem: %s (GUID: %s)", image.Path, guid)
		return c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "image_file_not_found",
		})
	}

	return c.File(image.Path)
}

func RegisterImageRoutes(api *echo.Group, handler *ImageHandler) {
	api.GET("/images/:guid", handler.GetImage)
}
