package grpc_users

import (
	"context"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/internal/logging"
	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	usersv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	grpc_middleware "github.com/DaanV2/mechanus/server/pkg/grpc/middleware"
	"github.com/DaanV2/mechanus/server/pkg/models/roles"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
)

var _ usersv1connect.UserServiceClient = &UserService{}

type UserService struct {
	users  *user_service.Service
	logger logging.Enriched
}

func NewUserService(users *user_service.Service) *UserService {
	logger := logging.Enriched{}.WithPrefix("grpc-users")

	return &UserService{
		users,
		logger,
	}
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
func (u *UserService) Get(ctx context.Context, req *connect.Request[usersv1.GetUserRequest]) (*connect.Response[usersv1.GetUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	jwt, err := grpc_middleware.JWTFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	r := roles.HighestRole(jwt.Claims.User.Roles)
	if r == roles.NONE {
		return nil, connect.NewError(connect.CodePermissionDenied, roles.ErrNotRightRole)
	}

	user, err := u.users.Get(req.Msg.Id)
	if err != nil {
		u.logger.From(ctx).Error("error during retrieve of user", "error", err, "id", req.Msg.Id)
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	msg := usersv1.GetUserResponse{
		User: &usersv1.User{
			Id: user.ID,
		},
	}

	if r.Inherits(roles.USER) {
		msg.User.Name = user.Username
	}

	return connect.NewResponse(&msg), nil
}
