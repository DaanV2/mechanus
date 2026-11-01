package grpc

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	usersv1 "github.com/DaanV2/mechanus/server/proto/users/v1"
	"github.com/DaanV2/mechanus/server/proto/users/v1/usersv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ usersv1connect.InviteServiceHandler = &InviteServiceHandler{}

// InviteServiceHandler lets admins hand out invite links, and new players create their account with one.
type InviteServiceHandler struct {
	invites *application.InviteService
	logger  logging.Enriched
}

func NewInviteServiceHandler(invites *application.InviteService) *InviteServiceHandler {
	return &InviteServiceHandler{
		invites: invites,
		logger:  logging.Enriched{}.WithPrefix("grpc-invites"),
	}
}

// Create implements usersv1connect.InviteServiceHandler.
func (h *InviteServiceHandler) Create(ctx context.Context, req *connect.Request[usersv1.CreateInviteRequest]) (*connect.Response[usersv1.CreateInviteResponse], error) {
	jwt, err := requireRole(ctx, roles.Admin)
	if err != nil {
		return nil, err
	}

	validFor := time.Duration(req.Msg.GetValidForSeconds()) * time.Second
	invite, err := h.invites.Create(ctx, jwt.Claims.User.ID, validFor)
	if err != nil {
		h.logger.From(ctx).Error("error creating invite", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	h.logger.From(ctx).Info("invite created", "inviteId", invite.ID, "expiresAt", invite.ExpiresAt)

	return connect.NewResponse(&usersv1.CreateInviteResponse{Invite: toInvite(invite)}), nil
}

// List implements usersv1connect.InviteServiceHandler.
func (h *InviteServiceHandler) List(ctx context.Context, _ *connect.Request[usersv1.ListInvitesRequest]) (*connect.Response[usersv1.ListInvitesResponse], error) {
	if _, err := requireRole(ctx, roles.Admin); err != nil {
		return nil, err
	}

	invites, err := h.invites.ListUsable(ctx)
	if err != nil {
		h.logger.From(ctx).Error("error listing invites", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*usersv1.Invite, 0, len(invites))
	for _, invite := range invites {
		result = append(result, toInvite(invite))
	}

	return connect.NewResponse(&usersv1.ListInvitesResponse{Invites: result}), nil
}

// Revoke implements usersv1connect.InviteServiceHandler.
func (h *InviteServiceHandler) Revoke(ctx context.Context, req *connect.Request[usersv1.RevokeInviteRequest]) (*connect.Response[usersv1.RevokeInviteResponse], error) {
	if _, err := requireRole(ctx, roles.Admin); err != nil {
		return nil, err
	}

	if err := h.invites.Revoke(ctx, req.Msg.GetId()); err != nil {
		if errors.Is(err, application.ErrInviteInvalid) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		h.logger.From(ctx).Error("error revoking invite", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.RevokeInviteResponse{}), nil
}

// Check implements usersv1connect.InviteServiceHandler.
func (h *InviteServiceHandler) Check(ctx context.Context, req *connect.Request[usersv1.CheckInviteRequest]) (*connect.Response[usersv1.CheckInviteResponse], error) {
	invite, err := h.invites.Get(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, application.ErrInviteInvalid) {
			return connect.NewResponse(&usersv1.CheckInviteResponse{Valid: false}), nil
		}
		h.logger.From(ctx).Error("error checking invite", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&usersv1.CheckInviteResponse{
		Valid:     true,
		ExpiresAt: timestamppb.New(invite.ExpiresAt),
	}), nil
}

// Accept implements usersv1connect.InviteServiceHandler.
func (h *InviteServiceHandler) Accept(ctx context.Context, req *connect.Request[usersv1.AcceptInviteRequest]) (*connect.Response[usersv1.AcceptInviteResponse], error) {
	username, password := req.Msg.GetUsername(), req.Msg.GetPassword()
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidUserPassword)
	}

	user, err := h.invites.Accept(ctx, req.Msg.GetId(), username, []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInviteInvalid):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, application.ErrUserAlreadyExists):
			return nil, connect.NewError(connect.CodeAlreadyExists, application.ErrUserAlreadyExists)
		case user != nil:
			// The account exists, only recording who used the invite failed
			h.logger.From(ctx).Warn("account created, but invite use not recorded", "error", err)
		default:
			h.logger.From(ctx).Error("error accepting invite", "error", err)

			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	h.logger.From(ctx).Info("account created with invite", "userId", user.ID)

	return connect.NewResponse(&usersv1.AcceptInviteResponse{User: &usersv1.User{Id: user.ID, Name: user.Username}}), nil
}

func toInvite(invite *models.Invite) *usersv1.Invite {
	return &usersv1.Invite{
		Id:        invite.ID,
		CreatedBy: invite.CreatedBy,
		CreatedAt: timestamppb.New(invite.CreatedAt),
		ExpiresAt: timestamppb.New(invite.ExpiresAt),
	}
}
