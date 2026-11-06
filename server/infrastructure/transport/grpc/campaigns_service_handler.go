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

const (
	CampaignLogger logging.Prefixed = "grpc-campaigns"
)

// CampaignServiceHandler implements the campaigns gRPC service.
type CampaignServiceHandler struct {
	campaigns *application.CampaignService
	users     *application.UserService
}

// NewCampaignServiceHandler creates a new CampaignServiceHandler.
func NewCampaignServiceHandler(campaigns *application.CampaignService, users *application.UserService) *CampaignServiceHandler {
	return &CampaignServiceHandler{
		campaigns: campaigns,
		users:     users,
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

	logger := CampaignLogger.From(ctx, "userId", jwt.Claims.User.ID, "campaignName", name)

	user, err := c.users.Get(ctx, jwt.Claims.User.ID)
	if err != nil {
		logger.Error("error during retrieve of user", "error", err)

		return nil, connect.NewError(connect.CodePermissionDenied, xerrors.ErrNotExist)
	}

	campaign := &models.Campaign{Name: name, Description: strings.TrimSpace(req.Msg.GetDescription())}
	err = c.campaigns.CreateWithUser(ctx, campaign, user)
	if err != nil {
		logger.Error("error during creation of campaign", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&campaignsv1.CreateCampaignResponse{Campaign: toCampaign(campaign, true)}), nil
}

// Get implements campaignsv1connect.CampaignServiceHandler.
func (c *CampaignServiceHandler) Get(ctx context.Context, req *connect.Request[campaignsv1.GetCampaignRequest]) (*connect.Response[campaignsv1.GetCampaignResponse], error) {
	jwt, err := requireRole(ctx, roles.Viewer)
	if err != nil {
		return nil, err
	}

	// Non-members get the same answer as for a missing campaign, so ids can't be probed
	campaign, err := c.visibleCampaign(ctx, jwt, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&campaignsv1.GetCampaignResponse{Campaign: toCampaign(campaign, canManage(jwt, campaign))}), nil
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
		CampaignLogger.From(ctx, "userId", jwt.Claims.User.ID).Error("error during listing of campaigns", "error", err)

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*campaignsv1.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		result = append(result, toCampaign(campaign, canManage(jwt, campaign)))
	}

	return connect.NewResponse(&campaignsv1.ListCampaignsResponse{Campaigns: result}), nil
}

func isMember(campaign *models.Campaign, userId string) bool {
	return slices.ContainsFunc(campaign.Users, func(u *models.User) bool {
		return u.ID == userId
	})
}

func toMember(user *models.User) *campaignsv1.CampaignMember {
	return &campaignsv1.CampaignMember{
		Id:   user.ID,
		Name: user.Username,
	}
}

func toMembers(users []*models.User) []*campaignsv1.CampaignMember {
	members := make([]*campaignsv1.CampaignMember, 0, len(users))
	for _, user := range users {
		members = append(members, toMember(user))
	}

	return members
}

func toCampaign(campaign *models.Campaign, manageable bool) *campaignsv1.Campaign {
	return &campaignsv1.Campaign{
		Id:          campaign.ID,
		Name:        campaign.Name,
		Description: campaign.Description,
		OwnerId:     campaign.OwnerID,
		CanManage:   manageable,
		CreatedAt:   timestamppb.New(campaign.CreatedAt),
		Members:     toMembers(campaign.Users),
	}
}
