package xerrors

import "errors"

var (
	// ErrNotExist is returned when an item does not exist.
	ErrNotExist = errors.New("item does not exist")
)
