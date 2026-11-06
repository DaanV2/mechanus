package models

// Campaign represents a game campaign in the database.
type Campaign struct {
	Model
	Name        string
	Description string
	// The user who created the campaign and may manage it, empty for campaigns made before owners were tracked
	OwnerID string  `gorm:"index"`
	Users   []*User `gorm:"many2many:user_campaigns"`
}
