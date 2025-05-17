package models

import "github.com/lib/pq"

type User struct {
	Model
	Name         string
	Roles        pq.StringArray `gorm:"type:text[]"`
	Campaigns    []*Campaign    `gorm:"many2many:user_campaigns"`
	Characters   []*Character   `gorm:"many2many:user_characters"`
	PasswordHash []byte
}

func (u *User) GetRoles() []string {
	return u.Roles
}

func (u *User) SetRoles(roles ...string) {
	u.Roles = roles
}