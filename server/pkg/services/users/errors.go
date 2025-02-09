package user_service

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)