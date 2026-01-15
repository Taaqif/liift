// Package repository provides database access patterns following the repository pattern.
package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepository provides common database operations for all repositories
type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepository) DBWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *BaseRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *BaseRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}
