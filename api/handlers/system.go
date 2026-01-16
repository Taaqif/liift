// Package handlers provides HTTP handlers
package handlers

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

const Version = "1.0.0"

type SystemInfoResponse struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func GetSystemInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, SystemInfoResponse{
		Version:   Version,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
}

func RegisterSystemRoutes(api *echo.Group) {
	api.GET("/system", GetSystemInfo)
}
