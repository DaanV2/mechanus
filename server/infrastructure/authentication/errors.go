package authentication

import "errors"

var (
	// ErrJTIRevoked is returned when a JWT ID has been revoked.
	ErrJTIRevoked = errors.New("JTI has been revoked")
	// ErrClaimsRead is returned when JWT claims cannot be read.
	ErrClaimsRead = errors.New("couldn't read claims on the jwt")
	ErrKIDMissing   = errors.New("couldn't find jwt kid header")
	ErrKIDNotString = errors.New("kid is not a string")
)
