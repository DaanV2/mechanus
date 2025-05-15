package models

type JTI struct {
	Model
	UserID  string `gorm:"index"`
	Revoked bool   `gorm:"default:false"`
}

func (j *JTI) Valid() bool {
	return j.ID != "" && !j.Revoked
}
