package database

import "errors"

var (
	ErrNotFound = errors.New("couldn't find specific entity") // Err: couldn't find specific entity
	ErrAlreadyExists = errors.New("entity already exists") // Err: entity already exists
)