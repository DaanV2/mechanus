package grpc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
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
	jwtStr := authentication.BearerFromHeaders(headers)
	if jwtStr == "" {
		return ctx
	}

	token, err := j.jwtService.Validate(ctx, jwtStr)
	if token == nil {
		// The token couldn't be parsed at all (malformed, expired, bad signature, ...), treat the request as anonymous
		logging.From(ctx).Debug("could not parse jwt", "error", err)

		return ctx
	}

	claims, ok := authentication.GetClaims(token.Claims)
	if !ok {
		j.logger.Error("somehow the claims are not expect as it should")

		return ctx
	}

	// Tokens with a limited scope (e.g. refresh only) can't be used to access the API
	valid := err == nil && claims.HasScope(authentication.ScopeAll)
	logger := logging.From(ctx).With("user.valid", valid, "user.id", claims.User.ID, "user.name", claims.User.Name)
	ctx = authentication.ContextWithJWT(ctx, claims, valid)
	ctx = logging.Context(ctx, logger)

	return ctx
}
