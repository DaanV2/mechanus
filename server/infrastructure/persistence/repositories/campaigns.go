package repositories

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

// CampaignRepository provides database operations for campaigns.
type CampaignRepository struct {
	db     *persistence.DB
	logger logging.Enriched
}

// NewCampaignRepository creates a new campaign repository with the provided database.
func NewCampaignRepository(db *persistence.DB) *CampaignRepository {
	return &CampaignRepository{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("campaigns"),
	}
}

// Get looks up the campaign by the given id, will return a [xerrors.ErrNotExist] if nothing matched.
func (repo *CampaignRepository) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	logger := repo.logger.With("campaignId", campaignId).From(ctx)
	logger.Debug("Getting campaign by id")

	var campaign models.Campaign

	tx := repo.db.WithContext(ctx).Preload("Users").First(&campaign, "id = ?", campaignId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &campaign, nil
}

// List returns all the campaigns, including their users, ordered by creation date.
func (repo *CampaignRepository) List(ctx context.Context) ([]*models.Campaign, error) {
	logger := repo.logger.From(ctx)
	logger.Debug("Listing all campaigns")

	var campaigns []*models.Campaign

	tx := repo.db.WithContext(ctx).Preload("Users").Order("created_at").Find(&campaigns)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return campaigns, nil
}

// ListForUser returns all the campaigns the given user is in, including their users, ordered by creation date.
func (repo *CampaignRepository) ListForUser(ctx context.Context, userId string) ([]*models.Campaign, error) {
	logger := repo.logger.With("userId", userId).From(ctx)
	logger.Debug("Listing campaigns for user")

	var campaigns []*models.Campaign

	tx := repo.db.WithContext(ctx).
		Preload("Users").
		Joins("JOIN user_campaigns ON user_campaigns.campaign_id = campaigns.id").
		Where("user_campaigns.user_id = ?", userId).
		Order("campaigns.created_at").
		Find(&campaigns)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return campaigns, nil
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

// LinkUser associates a user with a campaign.
func (repo *CampaignRepository) LinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := repo.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("add user to campaign")

	// Omit "Users.*" so only the join row is written: upserting the user would run its BeforeCreate hook, giving it a new ID
	return repo.db.WithContext(ctx).Model(campaign).Omit("Users.*").Association("Users").Append(user)
}

// UnlinkUser removes a user's association with a campaign.
func (repo *CampaignRepository) UnlinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := repo.logger.With("campaignId", campaign.ID, "userId", user.ID).From(ctx)
	logger.Debug("removing user to campaign")

	return repo.db.WithContext(ctx).Model(campaign).Association("Users").Delete(user)
}
