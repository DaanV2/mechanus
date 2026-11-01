package grpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
)

var (
	// ErrNotAuthenticated is returned when a request needs a user, but has no valid access token.
	// Clients should refresh their token or log in again.
	ErrNotAuthenticated = errors.New("missing, invalid or expired access token")
)

// requireAuthenticated returns the jwt of the request, or an Unauthenticated connect error when the request
// has no valid access token (none, malformed, expired, revoked or refresh-only).
func requireAuthenticated(ctx context.Context) (authentication.JWTContext, error) {
	jwt, err := authentication.JWTFromContext(ctx)
	if err != nil || !jwt.IsAuthenicated() {
		return jwt, connect.NewError(connect.CodeUnauthenticated, ErrNotAuthenticated)
	}

	return jwt, nil
}

// requireRole returns the jwt of the request, or a connect error:
// Unauthenticated when the request has no valid access token, PermissionDenied when the user lacks the role.
func requireRole(ctx context.Context, role roles.Role) (authentication.JWTContext, error) {
	jwt, err := requireAuthenticated(ctx)
	if err != nil {
		return jwt, err
	}

	if !jwt.IsAuthenicatedWithRole(role) {
		return jwt, connect.NewError(connect.CodePermissionDenied, errors.New("expecting atleast a "+role.String()+" role"))
	}

	return jwt, nil
}
