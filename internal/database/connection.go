// Package database provides database connection management.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB is the global database connection instance
	DB *gorm.DB
	// SQLDB is the underlying *sql.DB instance for connection pool management
	SQLDB *sql.DB
)

// Connect initializes the database connection with proper configuration
func Connect(config *Config) error {
	var dialector gorm.Dialector
	var err error

	switch config.Driver {
	case "sqlite":
		// Ensure the directory exists for SQLite database
		dbDir := filepath.Dir(config.SQLitePath)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}
		// SQLite will create the database file if it doesn't exist
		dialector = sqlite.Open(config.SQLitePath)
	case "postgres":
		dialector = postgres.Open(config.DSN())
	default:
		return fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// Configure GORM with optimized settings for production
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statement cache
	}

	// Open connection
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying *sql.DB for connection pool configuration
	SQLDB, err = DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for scalability
	SQLDB.SetMaxOpenConns(config.MaxOpenConns)
	SQLDB.SetMaxIdleConns(config.MaxIdleConns)
	SQLDB.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := SQLDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database (driver: %s)", config.Driver)
	return nil
}

// Close closes the database connection
func Close() error {
	if SQLDB != nil {
		return SQLDB.Close()
	}
	return nil
}

// Health checks the health of the database connection
func Health(ctx context.Context) error {
	if SQLDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return SQLDB.PingContext(ctx)
}
