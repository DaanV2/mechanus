package application

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
)

// CampaignService provides operations for managing campaigns in the application.
type CampaignService struct {
	repo *repositories.CampaignRepository
}

// NewCampaignService creates a new CampaignService with the provided repository.
func NewCampaignService(repo *repositories.CampaignRepository) *CampaignService {
	return &CampaignService{repo}
}

// Create creates a new campaign in the database.
func (s *CampaignService) Create(ctx context.Context, campaign *models.Campaign) error {
	return s.repo.Create(ctx, campaign)
}

// Get retrieves a campaign by its ID.
func (s *CampaignService) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	return s.repo.Get(ctx, campaignId)
}

// List retrieves all campaigns.
func (s *CampaignService) List(ctx context.Context) ([]*models.Campaign, error) {
	return s.repo.List(ctx)
}

// ListForUser retrieves all campaigns the given user is in.
func (s *CampaignService) ListForUser(ctx context.Context, userId string) ([]*models.Campaign, error) {
	return s.repo.ListForUser(ctx, userId)
}

// CreateWithUser creates a new campaign and adds the given user to it.
func (s *CampaignService) CreateWithUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	err := s.Create(ctx, campaign)
	if err != nil {
		return fmt.Errorf("creating campaign: %w", err)
	}

	err = s.AddUsers(ctx, campaign, user)
	if err != nil {
		return fmt.Errorf("adding user to campaign: %w", err)
	}

	return nil
}

// AddUsers adds a user to a campaign and updates the relationships in both objects.
func (s *CampaignService) AddUsers(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	err := s.repo.LinkUser(ctx, campaign, user)
	if err != nil {
		return err
	}

	campaign.Users = xslices.AddIfMissing(campaign.Users, user)
	user.Campaigns = xslices.AddIfMissing(user.Campaigns, campaign)

	return nil
}

// RemoveUser removes a user from a campaign and updates the relationships in both objects.
func (s *CampaignService) RemoveUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	err := s.repo.UnlinkUser(ctx, campaign, user)
	if err != nil {
		return err
	}

	campaign.Users = xslices.RemoveID(campaign.Users, user)
	user.Campaigns = xslices.RemoveID(user.Campaigns, campaign)

	return nil
}
