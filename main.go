package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"liift/api"
	"liift/internal/database"
	"liift/internal/utils"
	"liift/web"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database connection
	dbConfig, err := database.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Run database migrations
	if err := database.Migrate(database.DB); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Seed enum tables with valid values
	if err := database.SeedAll(database.DB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Create a new echo server
	e := echo.New()

	// Add standard middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Setup the web handlers to service vite static assets
	web.RegisterHandlers(e)

	// Setup the api handlers
	api.RegisterHandlers(e)

	// Start server in a goroutine
	port := utils.GetEnv("PORT", "3000")
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("Server exited")
}
