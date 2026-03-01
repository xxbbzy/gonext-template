package model

import (
	"time"

	"gorm.io/gorm"
)

// Item represents a CRUD example resource.
type Item struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"size:20;not null;default:active"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName overrides the table name.
func (Item) TableName() string {
	return "items"
}
