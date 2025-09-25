package models

type KeyValue struct {
	Key   string `gorm:"primaryKey;index"`
	Value []byte
}
