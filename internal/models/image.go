package models

import (
	"gorm.io/gorm"
)

// Image represents an uploaded image file
type Image struct {
	BaseModel
	GUID     string `gorm:"type:varchar(36);uniqueIndex;not null" json:"guid"`
	Path     string `gorm:"type:varchar(1000);not null" json:"path"`
	MimeType string `gorm:"type:varchar(100);not null" json:"mime_type"`
}

func (Image) TableName() string {
	return "images"
}

func (i *Image) BeforeCreate(tx *gorm.DB) error {
	if err := i.Validate(); err != nil {
		return err
	}
	return nil
}

func (i *Image) Validate() error {
	if i.GUID == "" {
		return gorm.ErrRecordNotFound
	}
	if i.Path == "" {
		return gorm.ErrRecordNotFound
	}
	if i.MimeType == "" {
		return gorm.ErrRecordNotFound
	}
	return nil
}
