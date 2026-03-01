package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Email        string         `json:"email" gorm:"size:100;not null;uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Role         string         `json:"role" gorm:"size:20;not null;default:user"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}
