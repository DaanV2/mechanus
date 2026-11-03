package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

// IdentityKind says who a connection turned out to be.
type IdentityKind int

const (
	// IdentityAnonymous is a connection without any credential.
	IdentityAnonymous IdentityKind = iota
	// IdentityUser is a connection with a valid user JWT.
	IdentityUser
	// IdentityDevice is a connection with a valid device token.
	IdentityDevice
)

func (k IdentityKind) String() string {
	switch k {
	case IdentityUser:
		return "user"
	case IdentityDevice:
		return "device"
	case IdentityAnonymous:
	}

	return "anonymous"
}

// ErrInvalidCredential is returned when a request carries a credential that isn't accepted: unknown scheme,
// bad signature, expired, revoked or meant for something else. It is never treated as no credential at all,
// otherwise a client can't tell "log in / pair again" from "not allowed".
var ErrInvalidCredential = errors.New("invalid credential")

type (
	// Identity is the outcome of resolving the credential of a connection, see [IdentityResolver].
	Identity struct {
		Kind IdentityKind
		// ID of the user or device, empty for anonymous
		ID string
		// The verified claims of the credential, nil for anonymous
		Claims *JWTClaims
		// The credential itself, without prefix, so long lived connections can check it again later
		Token string
		// When the credential expires, zero when it doesn't
		ExpiresAt time.Time
	}

	// IdentityResolver turns the credential of a request into an [Identity].
	IdentityResolver struct {
		jwts *JWTService
	}

	identity_context_key struct{}
)

// Anonymous returns the identity of a connection without a credential.
func Anonymous() *Identity {
	return &Identity{Kind: IdentityAnonymous}
}

// NewIdentityResolver creates a new IdentityResolver that validates credentials with the given service.
func NewIdentityResolver(jwts *JWTService) *IdentityResolver {
	return &IdentityResolver{jwts: jwts}
}

// Resolve returns who the request is. No credential gives [Anonymous], a credential that isn't accepted gives
// an error wrapping [ErrInvalidCredential]. A device token takes precedence over a user JWT, so a logged in browser
// can still act as the device it runs on. Authorization headers of an unknown scheme are invalid credentials.
func (r *IdentityResolver) Resolve(ctx context.Context, headers http.Header) (*Identity, error) {
	if token, ok := DeviceTokenFromHeaders(headers); ok {
		return r.resolveToken(ctx, token, IdentityDevice)
	}
	if hasUnknownAuthorization(headers) {
		return nil, fmt.Errorf("%w: unknown authorization scheme", ErrInvalidCredential)
	}
	if token := BearerFromHeaders(headers); token != "" {
		return r.resolveToken(ctx, token, IdentityUser)
	}

	return Anonymous(), nil
}

func (r *IdentityResolver) resolveToken(ctx context.Context, token string, kind IdentityKind) (*Identity, error) {
	t, err := r.jwts.Validate(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, err)
	}
	claims, ok := GetClaims(t.Claims)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, ErrClaimsRead)
	}

	// The scope keeps the two apart: a device token doesn't pass as a user, or the other way around.
	// Refresh tokens pass as neither.
	if !scopeAllows(claims, kind) {
		return nil, fmt.Errorf("%w: scope %q can't be used as a %s credential", ErrInvalidCredential, claims.Scope, kind)
	}

	identity := &Identity{Kind: kind, ID: claims.User.ID, Claims: claims, Token: token}
	if claims.ExpiresAt != nil {
		identity.ExpiresAt = claims.ExpiresAt.Time
	}
	logging.From(ctx).Debug("resolved identity", "kind", kind, "id", identity.ID)

	return identity, nil
}

func scopeAllows(claims *JWTClaims, kind IdentityKind) bool {
	switch kind {
	case IdentityUser:
		return claims.Scope == ScopeAll
	case IdentityDevice:
		return claims.Scope == ScopeDevice
	case IdentityAnonymous:
	}

	return false
}

// ContextWithIdentity adds the identity to the context.
func ContextWithIdentity(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identity_context_key{}, identity)
}

// IdentityFromContext returns the identity on the context, or [Anonymous] when there is none.
func IdentityFromContext(ctx context.Context) *Identity {
	if identity, ok := ctx.Value(identity_context_key{}).(*Identity); ok && identity != nil {
		return identity
	}

	return Anonymous()
}
