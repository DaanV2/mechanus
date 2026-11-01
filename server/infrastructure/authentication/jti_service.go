package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

// JTIService manages JWT IDs (JTI) in the system.
type JTIService struct {
	db     *persistence.DB
	logger logging.Enriched
}

// NewJTIService creates a new JTI service with the provided database.
func NewJTIService(db *persistence.DB) *JTIService {
	return &JTIService{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("jti_service"),
	}
}

// GetByUser retrieves all JTIs for a given user.
func (s *JTIService) GetByUser(ctx context.Context, userId string) ([]models.JTI, error) {
	if userId == "" {
		return nil, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("getting jti")

	var result []models.JTI
	tx := s.db.WithContext(ctx).Find(&result, "user_id = ?", userId)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return result, nil
}

// GetActive retrieves all active (non-revoked) JTIs for a given user.
func (s *JTIService) GetActive(ctx context.Context, userId string) ([]models.JTI, error) {
	if userId == "" {
		return nil, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("getting active jti's")

	var result []models.JTI
	// Fix: Use Where to filter by user_id and revoked = false
	tx := s.db.WithContext(ctx).Where("user_id = ? AND revoked = ?", userId, false).Find(&result)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return result, nil
}

// Get retrieves a JTI by its ID.
func (s *JTIService) Get(ctx context.Context, jti string) (*models.JTI, error) {
	if jti == "" {
		return nil, errors.New("string is empty: jti")
	}

	logger := s.logger.From(ctx).With("jti", jti)
	logger.Debug("getting jti")

	var result models.JTI
	tx := s.db.WithContext(ctx).First(&result, "id = ?", jti)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

// Create creates a new JTI for the given user.
func (s *JTIService) Create(ctx context.Context, userId string) (*models.JTI, error) {
	if userId == "" {
		return nil, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("creating jti")

	result := models.JTI{
		Model:   models.Model{},
		UserID:  userId,
		Revoked: false,
	}

	tx := s.db.WithContext(ctx).Create(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

// GetActiveOrCreate retrieves an active JTI for the user or creates one if none exist.
func (s *JTIService) GetActiveOrCreate(ctx context.Context, userId string) (*models.JTI, error) {
	v, err := s.GetActive(ctx, userId)
	if err == nil && len(v) > 0 {
		return &v[0], nil
	}

	return s.Create(ctx, userId)
}

// Revoke marks a JTI as revoked.
func (s *JTIService) Revoke(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, errors.New("string is empty: jti")
	}

	logger := s.logger.From(ctx).With("jti", jti)
	logger.Debug("revoking jti")

	tx := s.db.WithContext(ctx).Model(&models.JTI{}).Where("id = ?", jti).Update("revoked", true)
	if tx.Error != nil {
		return false, tx.Error
	}

	if tx.RowsAffected > 1 {
		logger.Error("revoked alot more then 1 JTI", "amount", tx.RowsAffected)
	}

	return tx.RowsAffected > 0, nil
}

// RevokeUser revokes all JTIs (sessions) of the given user, except the ones listed in except.
// It returns the amount of revoked JTIs.
func (s *JTIService) RevokeUser(ctx context.Context, userId string, except ...string) (int64, error) {
	if userId == "" {
		return 0, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("revoking all jti's of user")

	tx := s.db.WithContext(ctx).Model(&models.JTI{}).Where("user_id = ? AND revoked = ?", userId, false)
	if len(except) > 0 {
		tx = tx.Where("id NOT IN ?", except)
	}

	tx = tx.Update("revoked", true)
	if tx.Error != nil {
		return 0, tx.Error
	}

	return tx.RowsAffected, nil
}

// Touch marks the JTI as used now, so [JTIService.Cleanup] keeps it around.
func (s *JTIService) Touch(ctx context.Context, jti string) error {
	if jti == "" {
		return errors.New("string is empty: jti")
	}

	tx := s.db.WithContext(ctx).Model(&models.JTI{}).Where("id = ?", jti).Update("updated_at", time.Now())

	return tx.Error
}

// Cleanup permanently deletes revoked JTIs, and JTIs that haven't been used for longer than maxAge.
// It returns the amount of deleted JTIs.
func (s *JTIService) Cleanup(ctx context.Context, maxAge time.Duration) (int64, error) {
	logger := s.logger.From(ctx)

	cutoff := time.Now().Add(-maxAge)
	tx := s.db.WithContext(ctx).Unscoped().Where("revoked = ? OR updated_at < ?", true, cutoff).Delete(&models.JTI{})
	if tx.Error != nil {
		return 0, fmt.Errorf("error cleaning up jti's: %w", tx.Error)
	}

	logger.Debug("cleaned up jti's", "amount", tx.RowsAffected)

	return tx.RowsAffected, nil
}
