package jwts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	User  JWTUser `json:"user"`
	Scope string  `json:"scope"`
}

type JWTUser struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Roles     []string `json:"roles"`
	Campaigns []string `json:"campaigns"`
}

type JWTOptions struct {
	TokenDuration    time.Duration
	SigningAlgorithm string
}

type JWTService struct {
	options    *JWTOptions
	validator  *jwt.Validator
	jtiService *authenication.JTIService
	keys       JWKS
}

// TODO Refresh

func (s *JWTService) Create(ctx context.Context, user models.User, scope string) (string, error) {
	logging.From(ctx).Info("creating jwt")
	claims := &Claims{
		User: JWTUser{
			ID:        user.ID,
			Name:      user.Name,
			Roles:     user.Roles,
			Campaigns: user.Campaigns,
		},
		Scope: scope,
	}

	token, err := s.sign(claims)
	return token, err
}

func (s *JWTService) Validate(ctx context.Context, token string) (*jwt.Token, error) {
	jToken, err := s.validate(ctx, token, jwt.WithExpirationRequired(), jwt.WithIssuer(config.SERVICE_NAME))

	return jToken, err
}

func (s *JWTService) validate(ctx context.Context, token string, options ...jwt.ParserOption) (*jwt.Token, error) {
	logger := logging.From(ctx).With("jwt", token)
	logger.Debug("validating jwt token")

	jToken, err := jwt.ParseWithClaims(token, Claims{}, s.findKey, options...)
	if err != nil {
		logger.Error("jwt is not valid", "error", err)
		return nil, err
	}

	// Check the JTI is it has been revoked
	claims, ok := GetClaims(jToken.Claims)
	if !ok {
		return jToken, errors.New("couldn't read claims on the jwt")
	}

	jti, err := s.jtiService.Find(claims.User.ID, claims.ID)
	if err != nil {
		return jToken, fmt.Errorf("error finding the JTI: %w", err)
	}
	if jti.Revoked {
		return jToken, authenication.ErrJTIRevoked
	}

	return jToken, s.validator.Validate(jToken.Claims)
}

func (s *JWTService) sign(claims *Claims) (string, error) {
	// Get a token id
	jti, err := s.jtiService.GetOrCreate(claims.User.ID)
	if err != nil {
		return "", err
	}
	now := time.Now()

	expirationTime := now.Add(s.options.TokenDuration)
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		ID:        jti,
		Issuer:    config.SERVICE_NAME,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(time.Minute * -1)),
	}

	key, err := s.keys.GetSigningKey()
	if err != nil {
		return "", fmt.Errorf("trouble getting the signing key: %w", err)
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS512.Alg())
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = key.id

	strToken, err := token.SignedString(key.key)

	return strToken, err
}

func (s *JWTService) findKey(token *jwt.Token) (interface{}, error) {
	kidH, ok := token.Header["kid"]
	if !ok {
		return nil, errors.New("couldn't find jwt kid header")
	}

	kid, ok := kidH.(string)
	if !ok {
		return nil, errors.New("kid is not a string")
	}

	k := s.keys.GetKey(kid)
	if k != nil {
		return k, nil
	}

	return nil, errors.New("couldn't find the jwt signing key: " + kid)
}

func GetClaims(claims jwt.Claims) (Claims, bool) {
	if claims != nil {
		c, ok := claims.(Claims)
		if ok {
			return c, true
		}
	}

	return Claims{}, false
}
