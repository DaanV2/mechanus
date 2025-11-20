package models

import "github.com/lib/pq"

// User represents a user in the database.
type User struct {
	Model
	Username     string
	Roles        pq.StringArray `gorm:"type:text[]"`
	Campaigns    []*Campaign    `gorm:"many2many:user_campaigns"`
	Characters   []*Character   `gorm:"many2many:user_characters"`
	PasswordHash []byte
}

// GetRoles returns the user's roles.
func (u *User) GetRoles() []string {
	return u.Roles
}

// SetRoles sets the user's roles.
func (u *User) SetRoles(roles ...string) {
	u.Roles = roles
}
