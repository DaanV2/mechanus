package models

// KeyValue represents a key-value pair in the database for generic storage.
type KeyValue struct {
	Key   string `gorm:"primaryKey;index"`
	Value []byte
}
