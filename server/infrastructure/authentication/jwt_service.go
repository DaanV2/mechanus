package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/mechanus"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// JWT_ISSUER is the issuer name for JWT tokens.
	JWT_ISSUER   = mechanus.SERVICE_NAME
	JWT_AUDIENCE = mechanus.SERVICE_NAME
	// JWT_LEEWAY is the clock skew allowed when checking the time based claims (exp, nbf, iat).
	JWT_LEEWAY = time.Minute * 5
)

var _ lifecycle.AfterInitialize = &JWTService{}

type (
	// JWTOptions contains configuration options for JWT tokens.
	JWTOptions struct {
		// How long an access token ([ScopeAll]) is valid for.
		TokenDuration time.Duration
		// How long a refresh token ([ScopeRefresh]) is valid for.
		RefreshTokenDuration time.Duration
	}

	// JWTService manages JWT token creation and validation.
	JWTService struct {
		options    *JWTOptions
		jtiService *JTIService
		keys       *KeyManager
	}

	// TokenPair is the result of logging in or refreshing: an access token and the refresh token of the same session.
	TokenPair struct {
		Access  string
		Refresh string
	}
)

// DefaultJWTOptions returns the default token lifetimes.
func DefaultJWTOptions() JWTOptions {
	return JWTOptions{
		TokenDuration:        time.Hour,
		RefreshTokenDuration: time.Hour * 24 * 7,
	}
}

// NewJWTService creates a new JWT service with the provided JTI service and key manager.
func NewJWTService(jtiService *JTIService, keys *KeyManager) *JWTService {
	options := DefaultJWTOptions()
	service := &JWTService{
		options:    &options,
		jtiService: jtiService,
		keys:       keys,
	}

	return service
}

// WithOptions replaces the token options of the service.
func (s *JWTService) WithOptions(options JWTOptions) *JWTService {
	s.options = &options

	return s
}

// AfterInitialize implements lifecycle.AfterInitialize, it removes revoked and expired sessions.
func (s *JWTService) AfterInitialize(ctx context.Context) error {
	return s.Cleanup(ctx)
}

// Cleanup removes the JTIs of sessions that are revoked, or haven't been refreshed within the refresh token lifetime.
func (s *JWTService) Cleanup(ctx context.Context) error {
	_, err := s.jtiService.Cleanup(ctx, s.MaxTokenLifetime())

	return err
}

// MaxTokenLifetime returns the longest time any token issued by this service stays valid.
func (s *JWTService) MaxTokenLifetime() time.Duration {
	return max(s.options.TokenDuration, s.options.RefreshTokenDuration) + JWT_LEEWAY
}

// Create creates a new JWT token for the given user with the specified grant and scope, in a new session.
func (s *JWTService) Create(ctx context.Context, user *models.User, grant, scope string) (string, error) {
	logging.Info(ctx, "creating jwt")

	jti, err := s.jtiService.Create(ctx, user.ID)
	if err != nil {
		return "", err
	}

	return s.sign(ctx, jti.ID, newClaims(user, grant, scope))
}

// CreateSession starts a new session (JTI) for the user and returns its access and refresh token.
func (s *JWTService) CreateSession(ctx context.Context, user *models.User, grant string) (*TokenPair, error) {
	jti, err := s.jtiService.Create(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.signPair(ctx, jti.ID, user, grant)
}

// RefreshSession issues a new access and refresh token for an existing session, picking up the user's current roles.
func (s *JWTService) RefreshSession(ctx context.Context, user *models.User, jti string) (*TokenPair, error) {
	if err := s.jtiService.Touch(ctx, jti); err != nil {
		return nil, fmt.Errorf("error updating the session: %w", err)
	}

	return s.signPair(ctx, jti, user, GrantRefresh)
}

// Options returns the JWT service configuration options.
func (s *JWTService) Options() *JWTOptions {
	return s.options
}

// Validate validates a JWT token and returns the parsed token.
// It does not check the scope of the token, callers have to check [JWTClaims.HasScope] themselves.
func (s *JWTService) Validate(ctx context.Context, token string) (*jwt.Token, error) {
	return s.validate(ctx, token,
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(JWT_ISSUER),
		jwt.WithAudience(JWT_AUDIENCE),
		jwt.WithLeeway(JWT_LEEWAY),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS512.Alg()}),
	)
}

func (s *JWTService) validate(ctx context.Context, token string, options ...jwt.ParserOption) (*jwt.Token, error) {
	logger := logging.From(ctx)
	logger.Debug("validating jwt token")

	jToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, s.findPublicKeyFn(ctx), options...)
	if err != nil {
		logger.Debug("jwt is not valid", "error", err)

		return nil, err
	}

	// Check the JTI is it has been revoked
	claims, ok := GetClaims(jToken.Claims)
	if !ok {
		return jToken, ErrClaimsRead
	}

	logger = logger.With(
		"jti", claims.ID,
		"userId", claims.User.ID,
	)

	jti, err := s.jtiService.Get(ctx, claims.ID)
	if err != nil {
		return jToken, fmt.Errorf("error finding the JTI: %w", err)
	}

	if jti.UserID != claims.User.ID {
		logger.Error("a JTI has the wrong userId")

		return jToken, ErrJTIWrongUser
	}

	if jti.Revoked {
		return jToken, ErrJTIRevoked
	}

	return jToken, nil
}

func newClaims(user *models.User, grant, scope string) *JWTClaims {
	return &JWTClaims{
		User: JWTUser{
			ID:        user.ID,
			Name:      user.Username,
			Roles:     user.Roles,
			Campaigns: xslices.CollectIDs(user.Campaigns),
		},
		Grant: grant,
		Scope: scope,
	}
}

func (s *JWTService) signPair(ctx context.Context, jti string, user *models.User, grant string) (*TokenPair, error) {
	access, err := s.sign(ctx, jti, newClaims(user, grant, ScopeAll))
	if err != nil {
		return nil, err
	}

	refresh, err := s.sign(ctx, jti, newClaims(user, grant, ScopeRefresh))
	if err != nil {
		return nil, err
	}

	return &TokenPair{Access: access, Refresh: refresh}, nil
}

func (s *JWTService) sign(ctx context.Context, jti string, claims *JWTClaims) (string, error) {
	now := time.Now()

	duration := s.options.TokenDuration
	if claims.Scope == ScopeRefresh {
		duration = s.options.RefreshTokenDuration
	}

	expirationTime := now.Add(duration)
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		ID:        jti,
		Issuer:    JWT_ISSUER,
		Audience:  jwt.ClaimStrings{JWT_AUDIENCE},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(time.Minute * -1)),
		Subject:   claims.User.ID,
	}

	key, err := s.keys.GetSigningKey(ctx)
	if err != nil {
		return "", fmt.Errorf("trouble getting the signing key: %w", err)
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS512.Alg())
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = key.GetID()

	return token.SignedString(key.Private())
}

func (s *JWTService) findPublicKeyFn(ctx context.Context) func(token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		return s.findPublicKey(ctx, token)
	}
}

func (s *JWTService) findPublicKey(ctx context.Context, token *jwt.Token) (any, error) {
	kidH, ok := token.Header["kid"]
	if !ok {
		return nil, ErrKIDMissing
	}

	kid, ok := kidH.(string)
	if !ok {
		return nil, ErrKIDNotString
	}

	k, err := s.keys.Get(ctx, kid)
	if err != nil {
		return nil, err
	}

	if k != nil {
		return k.Public(), nil
	}

	return nil, errors.New("couldn't find the jwt signing key: " + kid)
}

// GetClaims extracts JWTClaims from generic jwt.Claims.
func GetClaims(claims jwt.Claims) (*JWTClaims, bool) {
	if claims != nil {
		c, ok := claims.(*JWTClaims)
		if ok {
			return c, true
		}
	}

	return nil, false
}
