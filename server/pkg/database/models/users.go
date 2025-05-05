package models

type User struct {
	Model
	Name         string
	Roles        []string
	Campaigns    []string
	PasswordHash []byte
}

func (u *User) GetRoles() []string {
	return u.Roles
}

func (u *User) SetRoles(roles ...string) {
	u.Roles = roles
}
