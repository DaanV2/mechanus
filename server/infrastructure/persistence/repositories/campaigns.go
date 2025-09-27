package repositories

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

type CampaignRepository struct {
	db     *persistence.DB
	logger logging.Enriched
}

func NewCampaignRepository(db *persistence.DB) *CampaignRepository {
	return &CampaignRepository{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("campaigns"),
	}
}

// Gets looks up the campaign by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (repo *CampaignRepository) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	logger := repo.logger.With("campaignId", campaignId).From(ctx)
	logger.Debug("Getting campaign by id")

	var campaign models.Campaign

	tx := repo.db.WithContext(ctx).First(&campaign, "id = ?", campaignId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &campaign, nil
}

// Create makes a new entry in the database
func (repo *CampaignRepository) Create(ctx context.Context, campaign *models.Campaign) error {
	logger := repo.logger.With("campaignname", campaign.Name).From(ctx)
	logger.Debug("Creating campaign")

	tx := repo.db.WithContext(ctx).Create(campaign)

	return tx.Error
}

// Update will take the new information in the campaign and update the database entry. Note, this does not update the password or the ID
func (repo *CampaignRepository) Update(ctx context.Context, campaign *models.Campaign) error {
	logger := repo.logger.With("campaignId", campaign.ID).From(ctx)
	logger.Debug("updating campaign")

	tx := repo.db.WithContext(ctx).Omit("id").Updates(campaign)

	return tx.Error
}

func (repo *CampaignRepository) LinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := repo.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("add user to campaign")

	return repo.db.WithContext(ctx).Model(campaign).Association("Users").Append(user)
}

func (repo *CampaignRepository) UnlinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := repo.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("removing user to campaign")

	return repo.db.WithContext(ctx).Model(campaign).Association("Users").Delete(user)
}
