package authentication

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/daanv2/go-kit/generics"
)

// JWTContext holds JWT authentication information in a context.
type JWTContext struct {
	Valid  bool
	Claims *JWTClaims
}

// IsAuthenicated checks if the JWT context represents an authenticated user.
func (j *JWTContext) IsAuthenicated() bool {
	return j.Valid && j.Claims != nil
}

// IsAuthenicatedWithRole checks if the user is authenticated and has the given role. Or a role that inherits it.
func (j *JWTContext) IsAuthenicatedWithRole(role roles.Role) bool {
	return j.IsAuthenicated() && roles.GrantsHasRole(j.Claims, role)
}

var (
	// ErrNoToken is returned when no JWT token is found in the context.
	ErrNoToken = errors.New("no token")
)

type jwt_context_key struct{}

// ContextWithJWT adds JWT claims to a context.
func ContextWithJWT(ctx context.Context, jwt *JWTClaims, valid bool) context.Context {
	c := JWTContext{
		Valid:  valid,
		Claims: jwt,
	}

	return context.WithValue(ctx, jwt_context_key{}, c)
}

// JWTFromContext returns the JWTContext and a ok whenever or not the request had a JWT token
func JWTFromContext(ctx context.Context) (JWTContext, error) {
	v := ctx.Value(jwt_context_key{})
	if v == nil {
		return generics.Empty[JWTContext](), ErrNoToken
	}

	c, ok := v.(JWTContext)
	if !ok {
		return c, ErrNoToken
	}

	return c, nil
}

// IsAuthenicated checks if the context contains an authenticated JWT.
func IsAuthenicated(ctx context.Context) bool {
	c, err := JWTFromContext(ctx)
	if err != nil {
		return false
	}

	return c.IsAuthenicated()
}

// IsAuthenicatedWithRole checks if the user is authenticated and has the given role. Or a role that inherits it.
func IsAuthenicatedWithRole(ctx context.Context, role roles.Role) bool {
	c, err := JWTFromContext(ctx)
	if err != nil {
		return false
	}

	return c.IsAuthenicatedWithRole(role)
}
