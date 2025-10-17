package grpc

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	usersv1 "github.com/DaanV2/mechanus/server/pkg/gen/proto/users/v1"
	"github.com/DaanV2/mechanus/server/pkg/gen/proto/users/v1/usersv1connect"
)

var _ usersv1connect.LoginServiceHandler = &LoginServiceHandler{}

var (
	ErrInvalidUserPassword = errors.New("username/password is invalid")
)

type LoginServiceHandler struct {
	users *application.UserService
	jwts  *authentication.JWTService
}

func NewLoginServiceHandler(users *application.UserService, jwts *authentication.JWTService) *LoginServiceHandler {
	return &LoginServiceHandler{
		users,
		jwts,
	}
}

// Login implements usersv1connect.LoginServiceHandler.
func (l *LoginServiceHandler) Login(ctx context.Context, req *connect.Request[usersv1.LoginRequest]) (*connect.Response[usersv1.LoginResponse], error) {
	// Check user
	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	u, err := l.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	ok, err := xcrypto.ComparePassword(u.PasswordHash, []byte(password))
	if err != nil || !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	// Valid user, create a token for them
	token, err := l.jwts.Create(ctx, u, "password")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("cannot create token: %w", err))
	}

	resp := connect.NewResponse(&usersv1.LoginResponse{Token: token, Type: "Bearer"})

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

	u, err := l.users.Get(ctx, claims.User.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	token, err = l.jwts.Create(ctx, u, "refresh")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := &usersv1.RefreshTokenResponse{Token: token, Type: "Bearer"}
	resp := connect.NewResponse(result)

	return resp, nil
}
