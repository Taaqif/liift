package main

import (
	"fmt"

	"liift/api"
	"liift/web"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create a new echo server
	e := echo.New()

	// Add standard middleware
	e.Use(middleware.RequestLogger())

	// Setup the web handlers to service vite static assets
	web.RegisterHandlers(e)

	// Setup the api handlers
	api.RegisterHandlers(e)

	// Start the server on port 3000
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 3000)))
}

