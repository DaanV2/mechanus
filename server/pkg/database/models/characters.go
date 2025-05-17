package models

type Character struct {
	Model
	Name  string
	Users []User `gorm:"many2many:user_characters"`
}
