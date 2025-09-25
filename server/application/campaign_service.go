package application

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
)

type CampaignService struct {
	db     *persistence.DB
	logger logging.Enriched
}

func NewCampaignService(db *persistence.DB) *CampaignService {
	return &CampaignService{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("campaigns"),
	}
}

// Gets looks up the campaign by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (s *CampaignService) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	logger := s.logger.With("campaignId", campaignId).From(ctx)
	logger.Debug("Getting campaign by id")

	var campaign models.Campaign

	tx := s.db.WithContext(ctx).First(&campaign, "id = ?", campaignId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &campaign, nil
}

// Create makes a new entry in the database
func (s *CampaignService) Create(ctx context.Context, campaign *models.Campaign) error {
	logger := s.logger.With("campaignname", campaign.Name).From(ctx)
	logger.Debug("Creating campaign")

	tx := s.db.WithContext(ctx).Create(campaign)

	return tx.Error
}

// Update will take the new information in the campaign and update the database entry. Note, this does not update the password or the ID
func (s *CampaignService) Update(ctx context.Context, campaign *models.Campaign) error {
	logger := s.logger.With("campaignId", campaign.ID).From(ctx)
	logger.Debug("updating campaign")

	tx := s.db.WithContext(ctx).Omit("id").Updates(campaign)

	return tx.Error
}

func (s *CampaignService) AddUsers(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := s.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("add user to campaign")

	campaign.Users = xslices.AddIfMissing(campaign.Users, user)
	user.Campaigns = xslices.AddIfMissing(user.Campaigns, campaign)

	return s.db.WithContext(ctx).Model(campaign).Association("Users").Append(user)
}

func (s *CampaignService) RemoveUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := s.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("removing user from campaign")

	campaign.Users = xslices.RemoveID(campaign.Users, user)
	user.Campaigns = xslices.RemoveID(user.Campaigns, campaign)

	return s.db.WithContext(ctx).Model(campaign).Association("Users").Delete(user)
}
