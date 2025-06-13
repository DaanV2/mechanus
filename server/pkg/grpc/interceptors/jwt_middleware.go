package grpc_interceptors

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/charmbracelet/log"
)

var _ connect.Interceptor = &JWTMiddleware{}

// JWTMiddleware checks incoming requests and validates the jwt if present
// The result is stored on the context and can be retrieved with [JWTFromContext]
type JWTMiddleware struct {
	jwtService *authenication.JWTService
	logger     *log.Logger
}

func NewJWTMiddleware(jwtService *authenication.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
		logger:     log.Default().WithPrefix("jwt middleware"),
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (j *JWTMiddleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next // dont do anything from client side
}

// WrapStreamingHandler implements connect.Interceptor.
func (j *JWTMiddleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx = j.validateAndInject(ctx, conn.RequestHeader())

		return next(ctx, conn)
	}
}

// WrapUnary implements connect.Interceptor.
func (j *JWTMiddleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		ctx = j.validateAndInject(ctx, ar.Header())

		return next(ctx, ar)
	}
}

// validateAndInject retrieves the bearer jwt from the request, validates it and stores the result on the context
// Which can be retrieved with [JWTFromContext]
func (j *JWTMiddleware) validateAndInject(ctx context.Context, headers http.Header) context.Context {
	auth := headers.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ctx
	}

	auth = strings.TrimPrefix(auth, "Bearer ")
	token, err := j.jwtService.Validate(ctx, auth)

	claims, ok := authenication.GetClaims(token.Claims)
	if !ok {
		c := JWTContext{
			Valid:  err == nil,
			Claims: claims,
		}
		ctx = ContextWithJWT(ctx, c)
		ctx = logging.Context(ctx, logging.From(ctx).With("user.valid", c.Claims, "user.id", c.Claims.User.ID, "user.name", c.Claims.User.Name))
	} else {
		j.logger.Error("somehow the claims are not expect as it should", "token", auth)
	}

	return ctx
}
