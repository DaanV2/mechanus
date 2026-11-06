package repositories

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"gorm.io/gorm"
)

const (
	CampaignLogger logging.Prefixed = "campaigns"
)

// CampaignRepository provides database operations for campaigns.
type CampaignRepository struct {
	db *persistence.DB
}

// NewCampaignRepository creates a new campaign repository with the provided database.
func NewCampaignRepository(db *persistence.DB) *CampaignRepository {
	return &CampaignRepository{
		db: db,
	}
}

// Get looks up the campaign by the given id, will return a [xerrors.ErrNotExist] if nothing matched.
func (repo *CampaignRepository) Get(ctx context.Context, campaignId string) (*models.Campaign, error) {
	logger := CampaignLogger.From(ctx, "campaignId", campaignId)
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
	logger := CampaignLogger.From(ctx)
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
	logger := CampaignLogger.From(ctx, "userId", userId)
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
	logger := CampaignLogger.From(ctx, "campaignname", campaign.Name)
	logger.Debug("Creating campaign")

	tx := repo.db.WithContext(ctx).Create(campaign)

	return tx.Error
}

// Update writes the name and description of the campaign to the database, empty values included.
// Other fields (owner, users, ...) are left alone.
func (repo *CampaignRepository) Update(ctx context.Context, campaign *models.Campaign) error {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID)
	logger.Debug("updating campaign")

	tx := repo.db.WithContext(ctx).
		Model(&models.Campaign{Model: models.Model{ID: campaign.ID}}).
		Updates(map[string]any{
			"name":        campaign.Name,
			"description": campaign.Description,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete soft deletes the campaign, its player links are kept so it can be restored.
func (repo *CampaignRepository) Delete(ctx context.Context, campaignId string) error {
	logger := CampaignLogger.From(ctx, "campaignId", campaignId)
	logger.Debug("deleting campaign")

	tx := repo.db.WithContext(ctx).Delete(&models.Campaign{}, "id = ?", campaignId)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// LinkUser associates a user with a campaign.
func (repo *CampaignRepository) LinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID, "userId", user.ID)
	logger.Debug("add user to campaign")

	// Omit "Users.*" so only the join row is written: upserting the user would run its BeforeCreate hook, giving it a new ID
	return repo.db.WithContext(ctx).Model(campaign).Omit("Users.*").Association("Users").Append(user)
}

// UnlinkUser removes a user's association with a campaign.
func (repo *CampaignRepository) UnlinkUser(ctx context.Context, campaign *models.Campaign, user *models.User) error {
	logger := CampaignLogger.From(ctx, "campaignId", campaign.ID, "userId", user.ID)
	logger.Debug("removing user to campaign")

	return repo.db.WithContext(ctx).Model(campaign).Association("Users").Delete(user)
}
