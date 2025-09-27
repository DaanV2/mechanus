package application

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
)

type CampaignService struct {
	repo *repositories.CampaignRepository
}

func NewCampaignService(repo *repositories.CampaignRepository) *CampaignService {
	return &CampaignService{repo}
}

func (s *CampaignService) Create(ctx context.Context, campaign *models.Campaign) error {
	return s.repo.Create(ctx, campaign)
}

func (s *CampaignService) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	return s.repo.Get(ctx, campaignId)
}

func (s *CampaignService) AddUsers(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	err := s.repo.LinkUser(ctx, campaign, user)
	if err != nil {
		return err
	}

	campaign.Users = xslices.AddIfMissing(campaign.Users, user)
	user.Campaigns = xslices.AddIfMissing(user.Campaigns, campaign)

	return nil
}

func (s *CampaignService) RemoveUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	err := s.repo.UnlinkUser(ctx, campaign, user)
	if err != nil {
		return err
	}

	campaign.Users = xslices.RemoveID(campaign.Users, user)
	user.Campaigns = xslices.RemoveID(user.Campaigns, campaign)

	return nil
}
