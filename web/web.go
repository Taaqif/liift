//go:build !dev

// Package web provides HTTP handlers for serving the frontend application.
package web

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	//go:embed dist/*
	dist embed.FS

	//go:embed dist/index.html
	indexHTML embed.FS

	distDirFS     = echo.MustSubFS(dist, "dist")
	distIndexHTML = echo.MustSubFS(indexHTML, "dist")
)

func RegisterHandlers(e *echo.Echo) {
	e.FileFS("/", "index.html", distIndexHTML)
	e.StaticFS("/", distDirFS)
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			return len(c.Path()) >= 4 && c.Path()[:4] == "/api"
		},
		Root:       "/",
		HTML5:      true,
		Browse:     false,
		IgnoreBase: true,
		Filesystem: http.FS(distDirFS),
	}))
}
