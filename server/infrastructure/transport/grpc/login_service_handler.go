package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
)

var _ usersv1connect.LoginServiceHandler = &LoginServiceHandler{}

var (
	ErrInvalidUserPassword = errors.New("username/password is invalid")
	// ErrTooManyLoginAttempts is returned when a username is locked out after too many failed logins.
	ErrTooManyLoginAttempts = errors.New("too many failed login attempts, try again later")
)

type LoginServiceHandler struct {
	users   *application.UserService
	jwts    *authentication.JWTService
	jtis    *authentication.JTIService
	limiter *loginLimiter
	logger  logging.Enriched
}

func NewLoginServiceHandler(users *application.UserService, jwts *authentication.JWTService, jtis *authentication.JTIService) *LoginServiceHandler {
	return &LoginServiceHandler{
		users:   users,
		jwts:    jwts,
		jtis:    jtis,
		limiter: newLoginLimiter(),
		logger:  logging.Enriched{}.WithPrefix("grpc-login"),
	}
}

// dummyPasswordHash is compared against when the username doesn't exist,
// so an unknown username takes as long as a wrong password and doesn't reveal which usernames exist.
var dummyPasswordHash = sync.OnceValue(func() []byte {
	hash, _ := xcrypto.HashPassword([]byte("mechanus-dummy-password"))

	return hash
})

// Login implements usersv1connect.LoginServiceHandler.
func (l *LoginServiceHandler) Login(ctx context.Context, req *connect.Request[usersv1.LoginRequest]) (*connect.Response[usersv1.LoginResponse], error) {
	// Check user
	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	if !l.limiter.Allowed(username) {
		l.logger.From(ctx).Warn("login rejected, too many failed attempts", "username", username)

		return nil, connect.NewError(connect.CodeResourceExhausted, ErrTooManyLoginAttempts)
	}

	u, err := l.users.FindByUsername(ctx, username)
	hash := dummyPasswordHash()
	if err == nil {
		hash = u.PasswordHash
	}

	ok, cerr := xcrypto.ComparePassword(hash, []byte(password))
	if err != nil || cerr != nil || !ok {
		l.limiter.Failed(username)

		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}
	l.limiter.Succeeded(username)

	// Valid user, start a new session for them
	tokens, err := l.jwts.CreateSession(ctx, u, authentication.GrantPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cannot create token: %w", err))
	}

	resp := connect.NewResponse(&usersv1.LoginResponse{Token: tokens.Access, Type: "Bearer", RefreshToken: tokens.Refresh})

	return resp, nil
}

// Refresh implements usersv1connect.LoginServiceHandler.
func (l *LoginServiceHandler) Refresh(ctx context.Context, req *connect.Request[usersv1.RefreshTokenRequest]) (*connect.Response[usersv1.RefreshTokenResponse], error) {
	token := req.Msg.GetToken()
	if token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing token"))
	}

	t, err := l.jwts.Validate(ctx, token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Get the user to update the info in th token
	claims, ok := authentication.GetClaims(t.Claims)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token not provided by mechanus"))
	}
	// Only refresh tokens can be refreshed, otherwise a leaked access token could be renewed forever
	if claims.Scope != authentication.ScopeRefresh {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only a refresh token can be refreshed"))
	}

	u, err := l.users.Get(ctx, claims.User.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user no longer exists"))
	}

	tokens, err := l.jwts.RefreshSession(ctx, u, claims.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := &usersv1.RefreshTokenResponse{Token: tokens.Access, Type: "Bearer", RefreshToken: tokens.Refresh}
	resp := connect.NewResponse(result)

	return resp, nil
}

// Logout implements usersv1connect.LoginServiceHandler.
func (l *LoginServiceHandler) Logout(ctx context.Context, req *connect.Request[usersv1.LogoutRequest]) (*connect.Response[usersv1.LogoutResponse], error) {
	var claims *authentication.JWTClaims

	if token := req.Msg.GetToken(); token != "" {
		t, err := l.jwts.Validate(ctx, token)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		c, ok := authentication.GetClaims(t.Claims)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token not provided by mechanus"))
		}
		claims = c
	} else {
		jwt, err := requireAuthenticated(ctx)
		if err != nil {
			return nil, err
		}
		claims = jwt.Claims
	}

	if _, err := l.jtis.Revoke(ctx, claims.ID); err != nil {
		l.logger.From(ctx).Error("error revoking session", "error", err, "jti", claims.ID)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	l.logger.From(ctx).Info("user logged out", "userId", claims.User.ID)

	return connect.NewResponse(&usersv1.LogoutResponse{}), nil
}
