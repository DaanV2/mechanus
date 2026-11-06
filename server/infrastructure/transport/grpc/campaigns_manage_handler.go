package grpc

import (
	"context"
	"errors"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xerrors"
	campaignsv1 "github.com/DaanV2/mechanus/server/proto/campaigns/v1"
)

// Who may do what with a campaign:
//   - admins and members can see a campaign and its players, everyone else gets "not found"
//   - admins and the owner (the user who created it) can edit, delete and add / remove players
//   - players can remove themselves (leave the campaign), except the owner, who can't be removed at all

var (
	// ErrNotCampaignOwner is returned when a member tries to manage a campaign they don't own.
	ErrNotCampaignOwner = errors.New("only the owner of the campaign or an admin can do this")
	// ErrInvalidPlayerUsername is returned when a player is added without a username.
	ErrInvalidPlayerUsername = errors.New("username cannot be empty")
	// ErrPlayerNotFound is returned when the user to add or remove doesn't exist or isn't a player.
	ErrPlayerNotFound = errors.New("player not found")
	// ErrAlreadyPlayer is returned when the user to add is already a player of the campaign.
	ErrAlreadyPlayer = errors.New("user is already a player of the campaign")
)

// Update implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) Update(ctx context.Context, req *connect.Request[campaignsv1.UpdateCampaignRequest]) (*connect.Response[campaignsv1.UpdateCampaignResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidCampaignName)
	}

	campaign, err := c.manageableCampaign(ctx, jwt, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	campaign.Name = name
	campaign.Description = strings.TrimSpace(req.Msg.GetDescription())

	return c.update(ctx, jwt, campaign)
}

// update saves the campaign, the caller has done the checks.
func (c *CampaignServiceHandler) update(ctx context.Context, jwt authentication.JWTContext, campaign *models.Campaign) (*connect.Response[campaignsv1.UpdateCampaignResponse], error) {
	if err := c.campaigns.Update(ctx, campaign); err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}
		CampaignLogger.From(ctx, "campaignId", campaign.ID).Error("error during update of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&campaignsv1.UpdateCampaignResponse{Campaign: toCampaign(campaign, canManage(jwt, campaign))}), nil
}

// Delete implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) Delete(ctx context.Context, req *connect.Request[campaignsv1.DeleteCampaignRequest]) (*connect.Response[campaignsv1.DeleteCampaignResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	campaign, err := c.manageableCampaign(ctx, jwt, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	return c.delete(ctx, jwt, campaign)
}

// delete soft deletes the campaign, the caller has done the checks.
func (c *CampaignServiceHandler) delete(ctx context.Context, jwt authentication.JWTContext, campaign *models.Campaign) (*connect.Response[campaignsv1.DeleteCampaignResponse], error) {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID, "userId", jwt.Claims.User.ID)

	if err := c.campaigns.Delete(ctx, campaign.ID); err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}
		logger.Error("error during delete of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Info("campaign deleted")

	return connect.NewResponse(&campaignsv1.DeleteCampaignResponse{}), nil
}

// ListPlayers implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) ListPlayers(ctx context.Context, req *connect.Request[campaignsv1.ListPlayersRequest]) (*connect.Response[campaignsv1.ListPlayersResponse], error) {
	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
	}

	campaign, err := c.visibleCampaign(ctx, jwt, req.Msg.GetCampaignId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&campaignsv1.ListPlayersResponse{Players: toMembers(campaign.Users)}), nil
}

// AddPlayer implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) AddPlayer(ctx context.Context, req *connect.Request[campaignsv1.AddPlayerRequest]) (*connect.Response[campaignsv1.AddPlayerResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	username := strings.TrimSpace(req.Msg.GetUsername())
	if username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidPlayerUsername)
	}

	campaign, err := c.manageableCampaign(ctx, jwt, req.Msg.GetCampaignId())
	if err != nil {
		return nil, err
	}

	user, err := c.users.FindByUsername(ctx, username)
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, ErrPlayerNotFound)
		}
		CampaignLogger.From(ctx, "campaignId", campaign.ID).Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if isMember(campaign, user.ID) {
		return nil, connect.NewError(connect.CodeAlreadyExists, ErrAlreadyPlayer)
	}

	return c.addPlayer(ctx, campaign, user)
}

// addPlayer links the user to the campaign, the caller has done the checks.
func (c *CampaignServiceHandler) addPlayer(ctx context.Context, campaign *models.Campaign, user *models.User) (*connect.Response[campaignsv1.AddPlayerResponse], error) {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID, "playerId", user.ID)

	if err := c.campaigns.AddUsers(ctx, campaign, user); err != nil {
		logger.Error("error during adding of player", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Info("player added to campaign")

	return connect.NewResponse(&campaignsv1.AddPlayerResponse{Player: toMember(user)}), nil
}

// RemovePlayer implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) RemovePlayer(ctx context.Context, req *connect.Request[campaignsv1.RemovePlayerRequest]) (*connect.Response[campaignsv1.RemovePlayerResponse], error) {
	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
	}

	campaign, err := c.visibleCampaign(ctx, jwt, req.Msg.GetCampaignId())
	if err != nil {
		return nil, err
	}

	playerId := req.Msg.GetUserId()
	// Players may leave a campaign themselves, everyone else needs to be able to manage it
	if playerId != jwt.Claims.User.ID && !canManage(jwt, campaign) {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrNotCampaignOwner)
	}

	idx := slices.IndexFunc(campaign.Users, func(u *models.User) bool { return u.ID == playerId })
	if idx < 0 {
		return nil, connect.NewError(connect.CodeNotFound, ErrPlayerNotFound)
	}

	return c.removePlayer(ctx, campaign, campaign.Users[idx])
}

// removePlayer unlinks the user from the campaign, the caller has done the checks.
func (c *CampaignServiceHandler) removePlayer(ctx context.Context, campaign *models.Campaign, user *models.User) (*connect.Response[campaignsv1.RemovePlayerResponse], error) {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID, "playerId", user.ID)

	if err := c.campaigns.RemoveUser(ctx, campaign, user); err != nil {
		if errors.Is(err, application.ErrCannotRemoveOwner) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		logger.Error("error during removal of player", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Info("player removed from campaign")

	return connect.NewResponse(&campaignsv1.RemovePlayerResponse{}), nil
}

// visibleCampaign looks up the campaign. It answers not found when the campaign doesn't exist, or when the caller is neither
// an admin nor a member, so campaign ids can't be probed.
func (c *CampaignServiceHandler) visibleCampaign(ctx context.Context, jwt authentication.JWTContext, id string) (*models.Campaign, error) {
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	campaign, err := c.campaigns.Get(ctx, id)
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}
		CampaignLogger.From(ctx, "campaignId", id).Error("error during retrieve of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !jwt.IsAuthenicatedWithRole(roles.Admin) && !isMember(campaign, jwt.Claims.User.ID) {
		return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
	}

	return campaign, nil
}

// manageableCampaign is [CampaignServiceHandler.visibleCampaign], but also requires the caller to be able to manage the campaign.
func (c *CampaignServiceHandler) manageableCampaign(ctx context.Context, jwt authentication.JWTContext, id string) (*models.Campaign, error) {
	campaign, err := c.visibleCampaign(ctx, jwt, id)
	if err != nil {
		return nil, err
	}
	if !canManage(jwt, campaign) {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrNotCampaignOwner)
	}

	return campaign, nil
}

// canManage tells whenever the caller may edit, delete and manage the players of the campaign: admins and the campaign owner.
func canManage(jwt authentication.JWTContext, campaign *models.Campaign) bool {
	if jwt.IsAuthenicatedWithRole(roles.Admin) {
		return true
	}

	return campaign.OwnerID != "" && campaign.OwnerID == jwt.Claims.User.ID
}
