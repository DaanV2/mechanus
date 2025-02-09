package grpc_users

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	usersv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	"github.com/DaanV2/mechanus/server/pkg/models/roles"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
)

var _ usersv1connect.UserServiceClient = &UserService{}

type UserService struct {
	users *user_service.Service
	jwts  *authenication.JWTService
}

// Create implements usersv1connect.UserServiceClient.
func (u *UserService) Create(ctx context.Context, req *connect.Request[usersv1.CreateAccountRequest]) (*connect.Response[usersv1.CreateAccountResponse], error) {
	username, password := req.Msg.Username, req.Msg.Password
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	user := users.User{
		Username:     username,
		PasswordHash: []byte(password),
		Roles:        []roles.Role{roles.USER},
		Campaigns:    []string{},
	}

	user, err := u.users.Create(user)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.CreateAccountResponse{User: &usersv1.User{Id: user.ID}}), nil
}

// Get implements usersv1connect.UserServiceClient.
func (u *UserService) Get(context.Context, *connect.Request[usersv1.GetUserRequest]) (*connect.Response[usersv1.GetUserResponse], error) {
	panic("unimplemented")
}
