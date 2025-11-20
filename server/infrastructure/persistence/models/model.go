package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model is the base model for all database entities with common fields.
type Model struct {
	ID        string `gorm:"primaryKey;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// GetID returns the model's unique identifier.
func (m *Model) GetID() string {
	return m.ID
}

// BeforeCreate is a GORM hook that generates a UUID before creating a record.
func (u *Model) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()

	return
}
