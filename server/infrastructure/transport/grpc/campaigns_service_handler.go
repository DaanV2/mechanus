package grpc

import (
	"context"
	"errors"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xerrors"
	campaignsv1 "github.com/DaanV2/mechanus/server/proto/campaigns/v1"
	"github.com/DaanV2/mechanus/server/proto/campaigns/v1/campaignsv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ campaignsv1connect.CampaignServiceHandler = &CampaignServiceHandler{}

var (
	// ErrInvalidCampaignName is returned when a campaign is created without a name.
	ErrInvalidCampaignName = errors.New("campaign name cannot be empty")
)

// CampaignServiceHandler implements the campaigns gRPC service.
type CampaignServiceHandler struct {
	campaigns *application.CampaignService
	users     *application.UserService
	logger    logging.Enriched
}

// NewCampaignServiceHandler creates a new CampaignServiceHandler.
func NewCampaignServiceHandler(campaigns *application.CampaignService, users *application.UserService) *CampaignServiceHandler {
	return &CampaignServiceHandler{
		campaigns: campaigns,
		users:     users,
		logger:    logging.Enriched{}.WithPrefix("grpc-campaigns"),
	}
}

// Create implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) Create(ctx context.Context, req *connect.Request[campaignsv1.CreateCampaignRequest]) (*connect.Response[campaignsv1.CreateCampaignResponse], error) {
	jwt, err := requireRole(ctx, roles.User)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidCampaignName)
	}

	logger := c.logger.With("userId", jwt.Claims.User.ID, "campaignName", name).From(ctx)

	user, err := c.users.Get(ctx, jwt.Claims.User.ID)
	if err != nil {
		logger.Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodePermissionDenied, xerrors.ErrNotExist)
	}

	campaign := &models.Campaign{Name: name}
	err = c.campaigns.CreateWithUser(ctx, campaign, user)
	if err != nil {
		logger.Error("error during creation of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&campaignsv1.CreateCampaignResponse{Campaign: toCampaign(campaign)}), nil
}

// Get implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) Get(ctx context.Context, req *connect.Request[campaignsv1.GetCampaignRequest]) (*connect.Response[campaignsv1.GetCampaignResponse], error) {
	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
	}

	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, xerrors.ErrNotExist)
	}

	logger := c.logger.With("campaignId", id, "userId", jwt.Claims.User.ID).From(ctx)

	campaign, err := c.campaigns.Get(ctx, id)
	if err != nil {
		if persistence.IsNotExist(err) {
			return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
		}
		logger.Error("error during retrieve of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Non-members get the same answer as for a missing campaign, so ids can't be probed
	if !jwt.IsAuthenicatedWithRole(roles.Admin) && !isMember(campaign, jwt.Claims.User.ID) {
		return nil, connect.NewError(connect.CodeNotFound, xerrors.ErrNotExist)
	}

	return connect.NewResponse(&campaignsv1.GetCampaignResponse{Campaign: toCampaign(campaign)}), nil
}

// List implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) List(ctx context.Context, _ *connect.Request[campaignsv1.ListCampaignsRequest]) (*connect.Response[campaignsv1.ListCampaignsResponse], error) {
	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
	}

	var campaigns []*models.Campaign
	if jwt.IsAuthenicatedWithRole(roles.Admin) {
		campaigns, err = c.campaigns.List(ctx)
	} else {
		campaigns, err = c.campaigns.ListForUser(ctx, jwt.Claims.User.ID)
	}
	if err != nil {
		c.logger.With("userId", jwt.Claims.User.ID).From(ctx).Error("error during listing of campaigns", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*campaignsv1.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		result = append(result, toCampaign(campaign))
	}

	return connect.NewResponse(&campaignsv1.ListCampaignsResponse{Campaigns: result}), nil
}

func isMember(campaign *models.Campaign, userId string) bool {
	return slices.ContainsFunc(campaign.Users, func(u *models.User) bool {
		return u.ID == userId
	})
}

func toCampaign(campaign *models.Campaign) *campaignsv1.Campaign {
	members := make([]*campaignsv1.CampaignMember, 0, len(campaign.Users))
	for _, user := range campaign.Users {
		members = append(members, &campaignsv1.CampaignMember{
			Id:   user.ID,
			Name: user.Username,
		})
	}

	return &campaignsv1.Campaign{
		Id:        campaign.ID,
		Name:      campaign.Name,
		CreatedAt: timestamppb.New(campaign.CreatedAt),
		Members:   members,
	}
}
