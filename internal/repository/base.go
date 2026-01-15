// Package repository provides database access patterns following the repository pattern.
package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB returns the database instance
func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

// DBWithContext returns a database instance with context
func (r *BaseRepository) DBWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Begin starts a new transaction
func (r *BaseRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

// Transaction executes a function within a database transaction
func (r *BaseRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}
