package models

// JTI represents a JWT ID in the database for token management.
type JTI struct {
	Model
	UserID  string `gorm:"index"`
	Revoked bool   `gorm:"default:false"`
}

// Valid checks if the JTI is valid (has an ID and is not revoked).
func (j *JTI) Valid() bool {
	return j.ID != "" && !j.Revoked
}
