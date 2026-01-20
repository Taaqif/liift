// Package repository provides database access patterns following the repository pattern.
package repository

import (
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
