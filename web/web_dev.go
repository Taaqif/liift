//go:build dev

// Package web provides HTTP handlers for serving the frontend application.
package web

import (
	"log"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RegisterHandlers(e *echo.Echo) {
	log.Println("Running in dev mode: proxying to Vite at http://localhost:5173")
	setupDevProxy(e)
}

func setupDevProxy(e *echo.Echo) {
	url, err := url.Parse("http://localhost:5173")
	if err != nil {
		log.Fatal(err)
	}
	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: url,
		},
	})
	e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
		Skipper: func(c echo.Context) bool {
			return len(c.Path()) >= 4 && c.Path()[:4] == "/api"
		},
	}))
}
