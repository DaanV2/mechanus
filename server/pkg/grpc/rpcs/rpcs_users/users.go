package rpcs_users

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/authenication/roles"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	usersv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1"
	"github.com/DaanV2/mechanus/server/pkg/grpc/gen/users/v1/usersv1connect"
	grpc_middleware "github.com/DaanV2/mechanus/server/pkg/grpc/middleware"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
)

var _ usersv1connect.UserServiceHandler = &UserService{}

type UserService struct {
	users  *user_service.Service
	logger logging.Enriched

	roleService *roles.RoleService
}

func NewUserService(users *user_service.Service) *UserService {
	logger := logging.Enriched{}.WithPrefix("grpc-users")
	roleService := &roles.RoleService{}

	return &UserService{
		users,
		logger,
		roleService,
	}
}

// Create implements usersv1connect.UserServiceClient.
func (u *UserService) Create(ctx context.Context, req *connect.Request[usersv1.CreateAccountRequest]) (*connect.Response[usersv1.CreateAccountResponse], error) {
	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	user := models.User{
		Name:         username,
		PasswordHash: []byte(password),
		Roles:        []string{"user"},
		Campaigns:    []*models.Campaign{},
	}

	err := u.users.Create(ctx, &user)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.CreateAccountResponse{User: &usersv1.User{Id: user.ID}}), nil
}

// Get implements usersv1connect.UserServiceClient.
func (u *UserService) Get(ctx context.Context, req *connect.Request[usersv1.GetUserRequest]) (*connect.Response[usersv1.GetUserResponse], error) {
	if req.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	id := req.Msg.GetId()
	logger := u.logger.With("userId", id)

	jwt, err := grpc_middleware.JWTFromContext(ctx)
	if err != nil {
		logger.From(ctx).Error("error during reading jwt", "error", err)

		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	if !u.roleService.HasRole(jwt.Claims, roles.Viewer) {
		logger.From(ctx).Error("user does not have the correct permissions", "roles", jwt.Claims.User.Roles)

		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("wrong permissions"))
	}

	user, err := u.users.Get(ctx, req.Msg.GetId())
	if err != nil {
		logger.From(ctx).Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	msg := usersv1.GetUserResponse{
		User: &usersv1.User{
			Id: user.ID,
		},
	}

	if u.roleService.HasRole(jwt.Claims, roles.User) {
		msg.User.Name = user.Name
	}

	return connect.NewResponse(&msg), nil
}
