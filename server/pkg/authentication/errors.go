package authentication

import "errors"

var (
	ErrJTIRevoked   = errors.New("JTI has been revoked")
	ErrClaimsRead   = errors.New("couldn't read claims on the jwt")
	ErrKIDMissing   = errors.New("couldn't find jwt kid header")
	ErrKIDNotString = errors.New("kid is not a string")
)
