package grpc

import (
	"context"
	"errors"
	"slices"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xerrors"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
)

var _ usersv1connect.UserServiceHandler = &UserServiceHandler{}

var (
	// ErrWrongPassword is returned when the current password given to change a password is wrong.
	ErrWrongPassword = errors.New("current password is wrong")
	// ErrNoRoles is returned when roles are set to an empty list.
	ErrNoRoles = errors.New("a user needs atleast one role")
	// ErrRemoveOwnAdmin is returned when an admin tries to remove their own admin role.
	ErrRemoveOwnAdmin = errors.New("you can't remove your own admin role")
)

// UserServiceOptions contains the configuration of the [UserServiceHandler].
type UserServiceOptions struct {
	// Whenever or not anyone can create an account, when false only admins can create accounts.
	SignupEnabled bool
}

type UserServiceHandler struct {
	users   *application.UserService
	jtis    *authentication.JTIService
	options UserServiceOptions
	logger  logging.Enriched
}

func NewUserServiceHandler(users *application.UserService, jtis *authentication.JTIService, options UserServiceOptions) *UserServiceHandler {
	return &UserServiceHandler{
		users:   users,
		jtis:    jtis,
		options: options,
		logger:  logging.Enriched{}.WithPrefix("grpc-users"),
	}
}

// Create implements usersv1connect.UserServiceClient.
func (u *UserServiceHandler) Create(ctx context.Context, req *connect.Request[usersv1.CreateAccountRequest]) (*connect.Response[usersv1.CreateAccountResponse], error) {
	if !u.options.SignupEnabled {
		if _, err := requireRole(ctx, roles.Admin); err != nil {
			return nil, err
		}
	}

	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidUserPassword)
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

	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
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
	if _, err := requireRole(ctx, roles.Admin); err != nil {
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
	if _, err := requireRole(ctx, roles.Admin); err != nil {
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

	// Sign out every session of the user, whoever knew the old password shouldn't stay logged in
	revoked, err := u.jtis.RevokeUser(ctx, id)
	if err != nil {
		logger.From(ctx).Error("error revoking sessions after password reset", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.From(ctx).Info("password reset by admin", "revokedSessions", revoked)

	return connect.NewResponse(&usersv1.ResetPasswordResponse{}), nil
}

// ChangePassword implements usersv1connect.UserServiceHandler.
func (u *UserServiceHandler) ChangePassword(ctx context.Context, req *connect.Request[usersv1.ChangePasswordRequest]) (*connect.Response[usersv1.ChangePasswordResponse], error) {
	jwt, err := requireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	current, password := req.Msg.GetCurrentPassword(), req.Msg.GetNewPassword()
	if password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, application.ErrEmptyPassword)
	}

	id := jwt.Claims.User.ID
	logger := u.logger.With("userId", id)

	user, err := u.users.Get(ctx, id)
	if err != nil {
		logger.From(ctx).Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
	}

	ok, err := xcrypto.ComparePassword(user.PasswordHash, []byte(current))
	if err != nil || !ok {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrWrongPassword)
	}

	if err := u.users.UpdatePassword(ctx, id, []byte(password)); err != nil {
		logger.From(ctx).Error("error during password change", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Keep the session that changed the password, sign out all the others
	revoked, err := u.jtis.RevokeUser(ctx, id, jwt.Claims.ID)
	if err != nil {
		logger.From(ctx).Error("error revoking sessions after password change", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.From(ctx).Info("password changed", "revokedSessions", revoked)

	return connect.NewResponse(&usersv1.ChangePasswordResponse{}), nil
}

// SetRoles implements usersv1connect.UserServiceHandler.
func (u *UserServiceHandler) SetRoles(ctx context.Context, req *connect.Request[usersv1.SetRolesRequest]) (*connect.Response[usersv1.SetRolesResponse], error) {
	jwt, err := requireRole(ctx, roles.Admin)
	if err != nil {
		return nil, err
	}

	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	newRoles := make([]string, 0, len(req.Msg.GetRoles()))
	for _, r := range req.Msg.GetRoles() {
		role, perr := roles.ParseRole(r)
		if perr != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, perr)
		}
		if !slices.Contains(newRoles, role.String()) {
			newRoles = append(newRoles, role.String())
		}
	}
	if len(newRoles) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNoRoles)
	}
	if id == jwt.Claims.User.ID && !slices.Contains(newRoles, roles.Admin.String()) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, ErrRemoveOwnAdmin)
	}

	logger := u.logger.With("userId", id)

	if _, err := u.users.Get(ctx, id); err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err := u.users.Update(ctx, &models.User{ID: id, Roles: newRoles}); err != nil {
		logger.From(ctx).Error("error updating roles", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	user, err := u.getFullInfo(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.From(ctx).Info("roles changed by admin", "roles", newRoles)

	return connect.NewResponse(&usersv1.SetRolesResponse{User: user}), nil
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
