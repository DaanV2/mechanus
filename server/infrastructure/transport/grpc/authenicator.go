package grpc

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/authentication"
	"github.com/charmbracelet/log"
)

var _ connect.Interceptor = &Authenticator{}

// Authenticator checks incoming requests and validates the jwt if present
// The result is stored on the context and can be retrieved with [authentication.JWTFromContext]
type Authenticator struct {
	jwtService *authentication.JWTService
	logger     *log.Logger
}

// NewAuthenticator builds a new Authenticator Interceptor
func NewAuthenticator(jwtService *authentication.JWTService) *Authenticator {
	return &Authenticator{
		jwtService: jwtService,
		logger:     log.Default().WithPrefix("jwt middleware"),
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (j *Authenticator) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		// TODO: See how we can inject the JWT into the request headers
		return next(ctx, s)
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (j *Authenticator) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx = j.validateAndInject(ctx, conn.RequestHeader())

		return next(ctx, conn)
	}
}

// WrapUnary implements connect.Interceptor.
func (j *Authenticator) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		ctx = j.validateAndInject(ctx, ar.Header())

		return next(ctx, ar)
	}
}

// validateAndInject retrieves the bearer jwt from the request, validates it and stores the result on the context
// Which can be retrieved with [authentication.JWTFromContext]
func (j *Authenticator) validateAndInject(ctx context.Context, headers http.Header) context.Context {
	jwtStr := getJwtValue(headers)
	if jwtStr == "" {
		return ctx
	}

	jwtStr = strings.TrimPrefix(jwtStr, "Bearer ")
	token, err := j.jwtService.Validate(ctx, jwtStr)

	claims, ok := authentication.GetClaims(token.Claims)
	if ok {
		logger := logging.From(ctx).With("user.valid", err == nil, "user.id", claims.User.ID, "user.name", claims.User.Name)
		ctx = authentication.ContextWithJWT(ctx, claims, err == nil)
		ctx = logging.Context(ctx, logger)
	} else {
		j.logger.Error("somehow the claims are not expect as it should", "token", jwtStr)
	}

	return ctx
}

func getJwtValue(headers http.Header) string {
	for _, auth := range headers.Values("Authorization") {
		if strings.HasPrefix(auth, "Bearer ") {
			return auth
		}
	}

	for _, cookieData := range headers.Values("Cookie") {
		cookies, err := http.ParseCookie(cookieData)
		if err != nil {
			continue
		}

		for _, cookie := range cookies {
			if cookie.Name == "access-token" && strings.HasPrefix(cookie.Value, "Bearer ") {
				return cookie.Value
			}
		}
	}

	return ""
}
