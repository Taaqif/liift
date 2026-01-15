// Package handlers provides HTTP handlers
package handlers

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

const Version = "1.0.0"

func GetSystemInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"version":    Version,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	})
}

func RegisterSystemRoutes(api *echo.Group) {
	api.GET("/system", GetSystemInfo)
}
