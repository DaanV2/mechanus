package authenication

import "errors"

var (
	ErrJTIRevoked = errors.New("JTI has been revoked")
)