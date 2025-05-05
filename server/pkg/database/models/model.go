package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Note: Gorm will fail if the function signature
//  does not include `*gorm.DB` and `error`

func (user *Model) BeforeCreate(tx *gorm.DB) error {
	// UUID version 4
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now()
	return nil
}

func (user *Model) BeforeSave(tx *gorm.DB) error {
	user.UpdatedAt = time.Now()
	return nil
}

func (user *Model) BeforeDelete(tx *gorm.DB) error {
	user.DeletedAt = gorm.DeletedAt{
		Time: time.Now(),
		Valid: true,
	}
	return nil
}