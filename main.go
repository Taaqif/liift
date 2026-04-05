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

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echolog "github.com/labstack/gommon/log"
)

func main() {
	// Load .env before anything reads environment variables.
	// In production (Docker) env vars are injected directly; this is a no-op.
	_ = godotenv.Load()

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

	if err := database.Migrate(database.DB); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	if err := database.SeedAll(database.DB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	jwtSecret := []byte(utils.MustGetEnv("JWT_SECRET"))

	e := echo.New()
	e.HideBanner = true

	logLevel := echolog.Lvl(utils.GetEnvAsInt("LOG_LEVEL", int(echolog.INFO)))
	e.Logger.SetLevel(logLevel)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	web.RegisterHandlers(e)
	api.RegisterHandlers(e, database.DB, jwtSecret)

	port := utils.GetEnv("PORT", "3000")
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("Server exited")
}
