package models

import "time"

// Invite is a single use link that lets someone create their own account.
// The ID is the secret part of the link. Revoking an invite soft deletes it.
type Invite struct {
	Model
	CreatedBy string `gorm:"index"`
	ExpiresAt time.Time
	// Set once the invite is claimed, so it can't be used twice
	UsedAt *time.Time
	// The account created with the invite
	UsedBy string
}

// Usable checks if the invite hasn't been used and hasn't expired at the given time.
func (i *Invite) Usable(now time.Time) bool {
	return i.ID != "" && i.UsedAt == nil && now.Before(i.ExpiresAt)
}
