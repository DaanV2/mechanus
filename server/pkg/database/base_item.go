package database

import (
	"time"

	"github.com/google/uuid"
)

type BaseItem struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func NewBaseItem() BaseItem {
	n := time.Now()
	return BaseItem{
		ID:        uuid.NewString(),
		CreatedAt: n,
		UpdatedAt: n,
		DeletedAt: nil,
	}
}

// IsDeleted returns true if this item has been marked as deleted
func (b BaseItem) IsDeleted() bool {
	return b.DeletedAt != nil
}

// Update will return a copy the date with the UpdatedAt field updated to now
func (b BaseItem) Update() BaseItem {
	b.UpdatedAt = time.Now()
	return b
}

// Update will return a copy the date with the UpdatedAt, DeletedAt field updated to now
func (b BaseItem) Delete() BaseItem {
	n := time.Now()
	b.UpdatedAt = n
	b.DeletedAt = &n
	return b
}