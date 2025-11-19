package models

// Campaign represents a game campaign in the database.
type Campaign struct {
	Model
	Name  string
	Users []*User `gorm:"many2many:user_campaigns"`
}
