package grpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

var _ connect.Interceptor = &PolicyInterceptor{}

// PolicyInterceptor guards the procedures of a service with a policy each. A credential that isn't accepted is rejected
// on every procedure, also on an anonymous one, and a procedure without a policy is denied.
// The identity is put on the context, see [authentication.IdentityFromContext].
type PolicyInterceptor struct {
	identities *authentication.IdentityResolver
	policies   map[string]authentication.Policy
}

// NewPolicyInterceptor creates a new PolicyInterceptor with the policy per procedure.
func NewPolicyInterceptor(identities *authentication.IdentityResolver, policies map[string]authentication.Policy) *PolicyInterceptor {
	return &PolicyInterceptor{identities, policies}
}

// WrapUnary implements connect.Interceptor.
func (p *PolicyInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := p.authorize(ctx, req.Spec().Procedure, req.Header())
		if err != nil {
			return nil, err
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (p *PolicyInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor.
func (p *PolicyInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := p.authorize(ctx, conn.Spec().Procedure, conn.RequestHeader())
		if err != nil {
			return err
		}

		return next(ctx, conn)
	}
}

func (p *PolicyInterceptor) authorize(ctx context.Context, procedure string, headers http.Header) (context.Context, error) {
	policy, ok := p.policies[procedure]
	if !ok {
		return ctx, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("no policy for %s", procedure))
	}

	identity, err := p.identities.Resolve(ctx, headers)
	if err != nil {
		logging.From(ctx).Info("rejected credential", "procedure", procedure, "error", err)

		return ctx, connect.NewError(connect.CodeUnauthenticated, authentication.ErrInvalidCredential)
	}

	if err := policy.Authorize(identity); err != nil {
		logging.From(ctx).Info("route policy denied", "procedure", procedure, "policy", policy, "identity", identity.Kind)
		if errors.Is(err, authentication.ErrIdentityRequired) {
			return ctx, connect.NewError(connect.CodeUnauthenticated, err)
		}

		return ctx, connect.NewError(connect.CodePermissionDenied, err)
	}

	return authentication.ContextWithIdentity(ctx, identity), nil
}
