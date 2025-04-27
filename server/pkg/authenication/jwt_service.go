package authenication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/constants"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	"github.com/golang-jwt/jwt/v5"
)

const (
	JWT_ISSUER   = constants.SERVICE_NAME
	JWT_AUDIENCE = constants.SERVICE_NAME
)

type (
	JWTOptions struct {
		TokenDuration time.Duration
	}

	JWTService struct {
		options    *JWTOptions
		validator  *jwt.Validator
		jtiService *JTIService
		keys       *KeyManager
	}
)

func NewJWTService(jtiService *JTIService, keys *KeyManager) *JWTService {
	service := &JWTService{
		options: &JWTOptions{
			TokenDuration: time.Hour * 1,
		},
		validator: jwt.NewValidator(
			jwt.WithAudience(JWT_AUDIENCE),
			jwt.WithIssuer(JWT_ISSUER),
			jwt.WithLeeway(time.Minute*5),
		),
		jtiService: jtiService,
		keys:       keys,
	}

	return service
}

// TODO Refresh

func (s *JWTService) Create(ctx context.Context, user users.User, scope string) (string, error) {
	logging.Info(ctx,"creating jwt")
	claims := &JWTClaims{
		User: JWTUser{
			ID:        user.ID,
			Name:      user.Username,
			Roles:     user.Roles,
			Campaigns: user.Campaigns,
		},
		Scope: scope,
	}

	return s.sign(claims)
}

func (s *JWTService) Validate(ctx context.Context, token string) (*jwt.Token, error) {
	jToken, err := s.validate(ctx, token, jwt.WithExpirationRequired(), jwt.WithIssuer(JWT_ISSUER))

	return jToken, err
}

func (s *JWTService) validate(ctx context.Context, token string, options ...jwt.ParserOption) (*jwt.Token, error) {
	logger := logging.From(ctx).With("jwt", token)
	logger.Debug("validating jwt token")

	jToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, s.findPublicKey, options...)
	if err != nil {
		logger.Error("jwt is not valid", "error", err)
		return nil, err
	}

	// Check the JTI is it has been revoked
	claims, ok := GetClaims(jToken.Claims)
	if !ok {
		return jToken, ErrClaimsRead
	}

	// Validate the token, then the JTI
	err = s.validator.Validate(jToken.Claims)
	if err != nil {
		return jToken, err
	}

	jti, err := s.jtiService.Find(claims.User.ID, claims.ID)
	if err != nil {
		return jToken, fmt.Errorf("error finding the JTI: %w", err)
	}
	if jti.Revoked {
		return jToken, ErrJTIRevoked
	}

	return jToken, nil
}

func (s *JWTService) sign(claims *JWTClaims) (string, error) {
	// Get a token id
	jti, err := s.jtiService.GetOrCreate(claims.User.ID)
	if err != nil {
		return "", fmt.Errorf("problem getting a jti: %w", err)
	}
	now := time.Now()

	expirationTime := now.Add(s.options.TokenDuration)
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		ID:        jti,
		Issuer:    JWT_ISSUER,
		Audience:  jwt.ClaimStrings{JWT_AUDIENCE},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(time.Minute * -1)),
	}

	key, err := s.keys.GetSigningKey()
	if err != nil {
		return "", fmt.Errorf("trouble getting the signing key: %w", err)
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS512.Alg())
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = key.ID()

	return token.SignedString(key.Private())
}

func (s *JWTService) findPublicKey(token *jwt.Token) (interface{}, error) {
	kidH, ok := token.Header["kid"]
	if !ok {
		return nil, ErrKIDMissing
	}

	kid, ok := kidH.(string)
	if !ok {
		return nil, ErrKIDNotString
	}

	k, err := s.keys.Get(kid)
	if err != nil {
		return nil, err
	}
	if k != nil {
		return k.Public(), nil
	}

	return nil, errors.New("couldn't find the jwt signing key: " + kid)
}

func GetClaims(claims jwt.Claims) (*JWTClaims, bool) {
	if claims != nil {
		c, ok := claims.(*JWTClaims)
		if ok {
			return c, true
		}
	}

	return nil, false
}
