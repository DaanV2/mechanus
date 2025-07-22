package authenication

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/authenication/roles"
	"github.com/daanv2/go-kit/generics"
)

type JWTContext struct {
	Valid  bool
	Claims *JWTClaims
}

func (j *JWTContext) IsAuthenicated() bool {
	return j.Valid && j.Claims != nil
}

// IsAuthenicatedWithRole checks if the user is authenticated and has the given role. Or a role that inherits it.
func (j *JWTContext) IsAuthenicatedWithRole(role roles.Role) bool {
	return j.IsAuthenicated() && roles.GrantsHasRole(j.Claims, role)
}

var (
	ErrNoToken = errors.New("no token")
)

type jwt_context_key struct{}

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
