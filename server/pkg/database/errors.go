package database

import (
	"errors"

	"gorm.io/gorm"
)

// IsNotExist checks if the error is a record not found error
func IsNotExist(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
