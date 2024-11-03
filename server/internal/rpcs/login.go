package rpcs

import (
	"context"
	"errors"

	connect "connectrpc.com/connect"
	v1 "github.com/DaanV2/mechanus/server/internal/grpc/users/v1"
	"github.com/DaanV2/mechanus/server/internal/grpc/users/v1/usersv1connect"
	jwts "github.com/DaanV2/mechanus/server/services/jwt"
	"github.com/DaanV2/mechanus/server/services/users"
	"github.com/charmbracelet/log"
)

type LoginType string

const (
	PASSWORD LoginType = "password"
	REFRESH  LoginType = "refresh"
)

func (l LoginType) String() string {
	return string(l)
}

var _ usersv1connect.LoginServiceHandler = &LoginRPC{}

type LoginRPC struct {
	userService users.Service
	jwtService  jwts.JWTService
}

func (rpc *LoginRPC) Login(ctx context.Context, request *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	// TODO ratelimit as a middleware

	if request.Msg.Username == "" || request.Msg.Password == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("expected password and username to be filled in"),
		)
	}

	// Find user
	user, err := rpc.userService.GetByUsername(request.Msg.Username)
	logger := log.With("name", user.Name, "id", user.ID)
	if err != nil {
		logger.Error("error during grabbing of user", "error", err)
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("wrong username or password"),
		)
	}

	// Compare passwords
	ok, err := users.ComparePassword(user.PasswordHash, []byte(request.Msg.Password))
	if err != nil {
		logger.Error("error during checking of the password", "error", err)
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("wrong username or password"),
		)
	}
	if !ok {
		logger.Warn("invalid password attempt")
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("wrong username or password"),
		)
	}

	// Valid, create JWT
	logger.Info("user logged in")
	token, err := rpc.signIn(ctx, user, PASSWORD)
	resp := &v1.LoginResponse{
		Token: token,
	}

	return connect.NewResponse[v1.LoginResponse](resp), nil
}

func (rpc *LoginRPC) Create(ctx context.Context, request *connect.Request[v1.CreateAccountRequest]) (*connect.Response[v1.CreateAccountResponse], error) {
	if request.Msg.Username == "" || request.Msg.Password == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("expected password and username to be filled in"),
		)
	}

	user := users.User{
		Name:         request.Msg.Username,
		Roles:        make([]string, 0),
		Campaigns:    make([]string, 0),
		PasswordHash: []byte(request.Msg.Password),
	}

	user, err := rpc.userService.Create(user)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			err,
		)
	}

	log.Info("created user", "id", user.ID, "name", user.Name)
	token, err := rpc.signIn(ctx, user, PASSWORD)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			err,
		)
	}

	resp := &v1.CreateAccountResponse{
		Token: token,
	}

	return connect.NewResponse[v1.CreateAccountResponse](resp), nil
}

// Refresh implements usersv1connect.LoginServiceHandler.
func (rpc *LoginRPC) Refresh(context.Context, *connect.Request[v1.RefreshTokenRequest]) (*connect.Response[v1.RefreshTokenResponse], error) {
	panic("unimplemented")

	//TODO extract user from token
}

func (rpc *LoginRPC) signIn(ctx context.Context, user users.User, t LoginType) (string, error) {
	log.Info("logging in user", "name", user.Name, "id", user.ID)

	return rpc.jwtService.Create(ctx, user, t.String())
}
