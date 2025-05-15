package authenication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/constants"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
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

func (s *JWTService) Create(ctx context.Context, user *models.User, scope string) (string, error) {
	logging.Info(ctx, "creating jwt")

	claims := &JWTClaims{
		User: JWTUser{
			ID:        user.ID,
			Name:      user.Name,
			Roles:     user.Roles,
			Campaigns: user.Campaigns,
		},
		Scope: scope,
	}

	jti, err := s.jtiService.GetActiveOrCreate(ctx, user.ID)
	if err != nil {
		return "", err
	}

	return s.sign(ctx, jti.ID, claims)
}

func (s *JWTService) Validate(ctx context.Context, token string) (*jwt.Token, error) {
	jToken, err := s.validate(ctx, token, jwt.WithExpirationRequired(), jwt.WithIssuer(JWT_ISSUER))

	return jToken, err
}

func (s *JWTService) validate(ctx context.Context, token string, options ...jwt.ParserOption) (*jwt.Token, error) {
	logger := logging.From(ctx)
	logger.Debug("validating jwt token", "jwt", token)

	jToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, s.findPublicKeyFn(ctx), options...)
	if err != nil {
		logger.Error("jwt is not valid", "error", err)
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

	// Validate the token, then the JTI
	err = s.validator.Validate(jToken.Claims)
	if err != nil {
		return jToken, err
	}

	jti, err := s.jtiService.Get(ctx, claims.ID)
	if err != nil {
		return jToken, fmt.Errorf("error finding the JTI: %w", err)
	}

	if jti.UserID != claims.User.ID {
		logger.Error("a JTI has the wrong userId")
		return jToken, errors.New("this token doesn't belong to the user")
	}

	if jti.Revoked {
		return jToken, ErrJTIRevoked
	}

	return jToken, nil
}

func (s *JWTService) sign(ctx context.Context, jti string, claims *JWTClaims) (string, error) {
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

	key, err := s.keys.GetSigningKey(ctx)
	if err != nil {
		return "", fmt.Errorf("trouble getting the signing key: %w", err)
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS512.Alg())
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = key.GetID()

	return token.SignedString(key.Private())
}

func (s *JWTService) findPublicKeyFn(ctx context.Context) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		return s.findPublicKey(ctx, token)
	}
}

func (s *JWTService) findPublicKey(ctx context.Context, token *jwt.Token) (interface{}, error) {
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

func GetClaims(claims jwt.Claims) (*JWTClaims, bool) {
	if claims != nil {
		c, ok := claims.(*JWTClaims)
		if ok {
			return c, true
		}
	}

	return nil, false
}
