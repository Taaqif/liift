package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Username           string   `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password           string   `gorm:"type:varchar(255);not null" json:"-"`
	Email              *string  `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Name               string   `gorm:"type:varchar(255)" json:"name"`
	DateOfBirth        string   `gorm:"type:varchar(10)" json:"date_of_birth"`
	Gender             string   `gorm:"type:varchar(20)" json:"gender"`
	WeightKg           *float64 `json:"weight_kg"`
	HeightCm           *float64 `json:"height_cm"`
	OnboardingComplete bool     `gorm:"default:false" json:"onboarding_complete"`
	Role               string   `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" && len(u.Password) < 60 {
		return u.SetPassword(u.Password)
	}
	return nil
}
