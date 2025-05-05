package models

type User struct {
	Model
	Name         string
	Roles        []string
	Campaigns    []string
	PasswordHash []byte
}