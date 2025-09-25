package campaign_service

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xslices "github.com/DaanV2/mechanus/server/pkg/extensions/slices"
)

type Service struct {
	db     *database.DB
	logger logging.Enriched
}

func NewService(db *database.DB) *Service {
	return &Service{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("campaigns"),
	}
}

// Gets looks up the campaign by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (s *Service) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
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
func (s *Service) Create(ctx context.Context, campaign *models.Campaign) error {
	logger := s.logger.With("campaignname", campaign.Name).From(ctx)
	logger.Debug("Creating campaign")

	tx := s.db.WithContext(ctx).Create(campaign)

	return tx.Error
}

// Update will take the new information in the campaign and update the database entry. Note, this does not update the password or the ID
func (s *Service) Update(ctx context.Context, campaign *models.Campaign) error {
	logger := s.logger.With("campaignId", campaign.ID).From(ctx)
	logger.Debug("updating campaign")

	tx := s.db.WithContext(ctx).Omit("id").Updates(campaign)

	return tx.Error
}

func (s *Service) AddUsers(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := s.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("add user to campaign")

	campaign.Users = xslices.AddIfMissing(campaign.Users, user)
	user.Campaigns = xslices.AddIfMissing(user.Campaigns, campaign)

	return s.db.WithContext(ctx).Model(campaign).Association("Users").Append(user)
}

func (s *Service) RemoveUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := s.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("removing user from campaign")

	campaign.Users = xslices.RemoveID(campaign.Users, user)
	user.Campaigns = xslices.RemoveID(user.Campaigns, campaign)

	return s.db.WithContext(ctx).Model(campaign).Association("Users").Delete(user)
}
