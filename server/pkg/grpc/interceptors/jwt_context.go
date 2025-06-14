package grpc_interceptors

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/daanv2/go-kit/generics"
)

var (
	ErrNoToken = errors.New("no token")
)

type JWTContext struct {
	Valid  bool
	Claims *authenication.JWTClaims
}

func (j *JWTContext) IsValid() bool {
	return j.Valid && j.Claims != nil
}

type jwt_context_key struct{}

func ContextWithJWT(ctx context.Context, jwt JWTContext) context.Context {
	return context.WithValue(ctx, jwt_context_key{}, jwt)
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
