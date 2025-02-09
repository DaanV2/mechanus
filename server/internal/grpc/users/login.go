package grpc_users

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	usersv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
)

var _ usersv1connect.LoginServiceHandler = &LoginService{}

var (
	ErrInvalidUserPassword = errors.New("username/password is invalid")
)

type LoginService struct {
	users *user_service.Service
	jwts  *authenication.JWTService
}

func NewLoginService(users *user_service.Service, jwts *authenication.JWTService) *LoginService {
	return &LoginService{
		users,
		jwts,
	}
}

// Login implements usersv1connect.LoginServiceHandler.
func (l *LoginService) Login(ctx context.Context, req *connect.Request[usersv1.LoginRequest]) (*connect.Response[usersv1.LoginResponse], error) {
	// Check user
	username, password := req.Msg.Username, req.Msg.Password
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	u, err := l.users.GetByUsername(username)
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

	return connect.NewResponse(&usersv1.LoginResponse{
		Token: token,
	}), nil
}

// Refresh implements usersv1connect.LoginServiceHandler.
func (l *LoginService) Refresh(ctx context.Context, req *connect.Request[usersv1.RefreshTokenRequest]) (*connect.Response[usersv1.RefreshTokenResponse], error) {
	token := req.Msg.Token
	if token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing token"))
	}

	t, err := l.jwts.Validate(ctx, token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Get the user to update the info in th token
	claims, ok := authenication.GetClaims(t.Claims)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token not provided by mechanus"))
	}

	u, err := l.users.Get(claims.User.ID)

	token, err = l.jwts.Create(ctx, u, "refresh")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.RefreshTokenResponse{Token: token}), nil
}
