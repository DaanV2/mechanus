package repositories

import (
	"context"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

// InviteRepository provides database operations for invites.
type InviteRepository struct {
	db     *persistence.DB
	logger logging.Enriched
}

// NewInviteRepository creates a new invite repository with the provided database.
func NewInviteRepository(db *persistence.DB) *InviteRepository {
	return &InviteRepository{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("invites"),
	}
}

// Create stores a new invite, setting its ID.
func (repo *InviteRepository) Create(ctx context.Context, invite *models.Invite) error {
	repo.logger.With("createdBy", invite.CreatedBy).From(ctx).Debug("creating invite")

	return repo.db.WithContext(ctx).Create(invite).Error
}

// Get looks up the invite by the given id, revoked invites are not found.
func (repo *InviteRepository) Get(ctx context.Context, id string) (*models.Invite, error) {
	repo.logger.With("inviteId", id).From(ctx).Debug("getting invite")

	var invite models.Invite
	tx := repo.db.WithContext(ctx).First(&invite, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &invite, nil
}

// ListUsable returns the invites that are not used, revoked or expired at the given time.
func (repo *InviteRepository) ListUsable(ctx context.Context, now time.Time) ([]*models.Invite, error) {
	repo.logger.From(ctx).Debug("listing invites")

	var invites []*models.Invite
	tx := repo.db.WithContext(ctx).
		Where("used_at IS NULL AND expires_at > ?", now).
		Order("created_at").
		Find(&invites)

	return invites, tx.Error
}

// Claim marks the invite as used when it is still usable at the given time.
// Returns false when it was already used, revoked or expired, so concurrent claims only succeed once.
func (repo *InviteRepository) Claim(ctx context.Context, id string, now time.Time) (bool, error) {
	repo.logger.With("inviteId", id).From(ctx).Debug("claiming invite")

	tx := repo.db.WithContext(ctx).
		Model(&models.Invite{}).
		Where("id = ? AND used_at IS NULL AND expires_at > ?", id, now).
		Update("used_at", now)

	return tx.RowsAffected == 1, tx.Error
}

// Release undoes a claim that didn't lead to an account, so the invite can be used again.
func (repo *InviteRepository) Release(ctx context.Context, id string) error {
	repo.logger.With("inviteId", id).From(ctx).Debug("releasing invite")

	return repo.db.WithContext(ctx).
		Model(&models.Invite{}).
		Where("id = ?", id).
		Update("used_at", nil).Error
}

// SetUsedBy records the account that was created with the invite.
func (repo *InviteRepository) SetUsedBy(ctx context.Context, id, userID string) error {
	return repo.db.WithContext(ctx).
		Model(&models.Invite{}).
		Where("id = ?", id).
		Update("used_by", userID).Error
}

// Revoke soft deletes the invite, returns false when there was no such invite.
func (repo *InviteRepository) Revoke(ctx context.Context, id string) (bool, error) {
	repo.logger.With("inviteId", id).From(ctx).Debug("revoking invite")

	tx := repo.db.WithContext(ctx).Delete(&models.Invite{}, "id = ?", id)

	return tx.RowsAffected == 1, tx.Error
}
