package models

type Campaign struct {
	Model
	Name  string
	Users []User `gorm:"many2many:user_campaigns"`
}
