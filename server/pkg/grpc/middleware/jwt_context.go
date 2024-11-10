package grpc_middleware

import (
	"context"

	"github.com/DaanV2/mechanus/server/pkg/authenication"
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

func JWTFromContext(ctx context.Context) (JWTContext, bool) {
	var empty JWTContext
	v := ctx.Value(jwt_context_key{})
	if v == nil {
		return empty, false
	}

	c, ok := v.(JWTContext)
	return c, ok
}
