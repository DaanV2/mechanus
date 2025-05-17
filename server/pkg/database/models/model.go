package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"primaryKey;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) GetID() string {
	return m.ID
}

func (u *Model) BeforeCreate(tx *gorm.DB) (err error) {
  u.ID = uuid.NewString()
  return
}