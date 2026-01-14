package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MessageHandler handles the /api/message endpoint
func MessageHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hello, from the golang World!"})
}
