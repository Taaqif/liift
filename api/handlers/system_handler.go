// Package handlers provides HTTP handlers
package handlers

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

// Build-time variables — injected via -ldflags during go build.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

type SystemInfoResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func GetSystemInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, SystemInfoResponse{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
}

func RegisterSystemRoutes(api *echo.Group) {
	api.GET("/system", GetSystemInfo)
}
