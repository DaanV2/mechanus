package models

// Character represents a game character in the database.
type Character struct {
	Model
	Name  string
	Users []User `gorm:"many2many:user_characters"`
}
