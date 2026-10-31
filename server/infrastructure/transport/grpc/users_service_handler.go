package grpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xerrors"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
)

var _ usersv1connect.UserServiceHandler = &UserServiceHandler{}

type UserServiceHandler struct {
	users  *application.UserService
	logger logging.Enriched

	roleService *roles.RoleService
}

func NewUserServiceHandler(users *application.UserService) *UserServiceHandler {
	logger := logging.Enriched{}.WithPrefix("grpc-users")
	roleService := &roles.RoleService{}

	return &UserServiceHandler{
		users,
		logger,
		roleService,
	}
}

// Create implements usersv1connect.UserServiceClient.
func (u *UserServiceHandler) Create(ctx context.Context, req *connect.Request[usersv1.CreateAccountRequest]) (*connect.Response[usersv1.CreateAccountResponse], error) {
	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrInvalidUserPassword)
	}

	user := models.User{
		Username:     username,
		PasswordHash: []byte(password),
		Roles:        []string{roles.User.String()},
		Campaigns:    []*models.Campaign{},
	}

	err := u.users.Create(ctx, &user)
	if err != nil {
		if errors.Is(err, application.ErrUserAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.CreateAccountResponse{User: &usersv1.User{Id: user.ID}}), nil
}

// Get implements usersv1connect.UserServiceClient.
func (u *UserServiceHandler) Get(ctx context.Context, req *connect.Request[usersv1.GetUserRequest]) (*connect.Response[usersv1.GetUserResponse], error) {
	if req.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	id := req.Msg.GetId()
	logger := u.logger.With("userId", id)

	jwt, err := authentication.JWTFromContext(ctx)
	if err != nil {
		logger.From(ctx).Error("error during reading jwt", "error", err)

		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	if !jwt.IsAuthenicatedWithRole(roles.Viewer) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("expecting atleast a viewer role"))
	}

	// If admin or current user
	if jwt.IsAuthenicatedWithRole(roles.Admin) || jwt.Claims.User.ID == id {
		user, cerr := u.getFullInfo(ctx, id)
		if cerr != nil {
			logger.From(ctx).Error("error during retrieve of user", "error", cerr)

			return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
		}

		return connect.NewResponse(&usersv1.GetUserResponse{User: user}), nil
	}

	user, err := u.getReducedInfo(ctx, id)
	if err != nil {
		logger.From(ctx).Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	return connect.NewResponse(&usersv1.GetUserResponse{User: user}), nil
}

// List implements usersv1connect.UserServiceClient.
func (u *UserServiceHandler) List(ctx context.Context, _ *connect.Request[usersv1.ListUsersRequest]) (*connect.Response[usersv1.ListUsersResponse], error) {
	if err := u.requireAdmin(ctx); err != nil {
		return nil, err
	}

	users, err := u.users.Find(ctx, &models.User{})
	if err != nil {
		u.logger.From(ctx).Error("error during listing of users", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*usersv1.User, 0, len(users))
	for _, user := range users {
		result = append(result, toFullUser(user))
	}

	return connect.NewResponse(&usersv1.ListUsersResponse{Users: result}), nil
}

// ResetPassword implements usersv1connect.UserServiceClient.
func (u *UserServiceHandler) ResetPassword(ctx context.Context, req *connect.Request[usersv1.ResetPasswordRequest]) (*connect.Response[usersv1.ResetPasswordResponse], error) {
	if err := u.requireAdmin(ctx); err != nil {
		return nil, err
	}

	id, password := req.Msg.GetId(), req.Msg.GetNewPassword()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}
	if password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, application.ErrEmptyPassword)
	}

	logger := u.logger.With("userId", id)

	err := u.users.UpdatePassword(ctx, id, []byte(password))
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}
		logger.From(ctx).Error("error during password reset", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.From(ctx).Info("password reset by admin")

	return connect.NewResponse(&usersv1.ResetPasswordResponse{}), nil
}

func (u *UserServiceHandler) requireAdmin(ctx context.Context) error {
	jwt, err := authentication.JWTFromContext(ctx)
	if err != nil {
		u.logger.From(ctx).Error("error during reading jwt", "error", err)

		return connect.NewError(connect.CodePermissionDenied, err)
	}

	if !jwt.IsAuthenicatedWithRole(roles.Admin) {
		return connect.NewError(connect.CodePermissionDenied, errors.New("expecting an admin role"))
	}

	return nil
}

func toFullUser(user *models.User) *usersv1.User {
	return &usersv1.User{
		Id:    user.ID,
		Name:  user.Username,
		Roles: user.Roles,
	}
}

func (u *UserServiceHandler) getFullInfo(ctx context.Context, id string) (*usersv1.User, error) {
	user, err := u.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return toFullUser(user), nil
}

func (u *UserServiceHandler) getReducedInfo(ctx context.Context, id string) (*usersv1.User, error) {
	user, err := u.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &usersv1.User{
		Id:   user.ID,
		Name: user.Username,
	}, nil

}
