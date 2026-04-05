// Package database provides database configuration and connection management.
package database

import (
	"fmt"

	"liift/internal/utils"
)

// Config holds database configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int    // in seconds
	Driver       string // "postgres" or "sqlite"
	SQLitePath   string // path for SQLite database file
}

// LoadConfig loads database configuration from environment variables.
// godotenv.Load() must be called before this (done in main).
func LoadConfig() (*Config, error) {
	config := &Config{
		Host:         utils.GetEnv("DB_HOST", "localhost"),
		Port:         utils.GetEnvAsInt("DB_PORT", 5432),
		User:         utils.GetEnv("DB_USER", "postgres"),
		Password:     utils.GetEnv("DB_PASSWORD", ""),
		DBName:       utils.GetEnv("DB_NAME", "liift"),
		SSLMode:      utils.GetEnv("DB_SSLMODE", "disable"),
		MaxOpenConns: utils.GetEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: utils.GetEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		MaxLifetime:  utils.GetEnvAsInt("DB_MAX_LIFETIME", 300),
		Driver:       utils.GetEnv("DB_DRIVER", "postgres"),
		SQLitePath:   utils.GetEnv("DB_SQLITE_PATH", "./data/liift.db"),
	}

	return config, nil
}

// DSN returns the data source name string for PostgreSQL
func (c *Config) DSN() string {
	if c.Driver == "sqlite" {
		return c.SQLitePath
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
