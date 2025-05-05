package authenication

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
)

type JTIService struct {
	db     *database.DB
	logger logging.Enriched
}

func NewJTIService(db *database.DB) *JTIService {
	return &JTIService{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("jti_service"),
	}
}

func (s *JTIService) GetByUser(ctx context.Context, userId string) ([]models.JTI, error) {
	if userId == "" {
		return nil, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("getting jti")

	var result []models.JTI
	tx := s.db.WithContext(ctx).Find(&result, userId)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return result, nil
}

func (s *JTIService) GetActive(ctx context.Context, userId string) ([]models.JTI, error) {
	if userId == "" {
		return nil, errors.New("string is empty: userId")
	}

	logger := s.logger.From(ctx).With("userId", userId)
	logger.Debug("getting active jti's")

	var result []models.JTI
	tx := s.db.WithContext(ctx).Find(&result, "id = ?", userId, "revoked = ", false)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return result, nil
}

func (s *JTIService) Get(ctx context.Context, jti string) (*models.JTI, error) {
	if jti == "" {
		return nil, errors.New("string is empty: jti")
	}

	logger := s.logger.From(ctx).With("jti", jti)
	logger.Debug("getting jti")

	var result models.JTI
	tx := s.db.WithContext(ctx).First(&result, jti)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

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
	tx := s.db.WithContext(ctx).Create(result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}

// GetOrCreate is optimistic, and epics that all gets will almost always succeed
func (s *JTIService) GetActiveOrCreate(ctx context.Context, userId string) (*models.JTI, error) {
	v, err := s.GetActive(ctx, userId)
	if err == nil && len(v) > 0 {
		return &v[0], nil
	}

	return s.Create(ctx, userId)
}

func (s *JTIService) Revoke(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, errors.New("string is empty: jti")
	}

	logger := s.logger.From(ctx).With("jti", jti)
	logger.Debug("revoking jti")

	result := models.JTI{
		Model: models.Model{
			ID: jti,
		},
		Revoked: true,
	}
	tx := s.db.WithContext(ctx).Create(result)
	if tx.Error != nil {
		return false, tx.Error
	}
	if tx.RowsAffected > 1 {
		logger.Error("revoked alot more then 1 JTI", "amount", tx.RowsAffected)
	}

	return tx.RowsAffected > 0, nil
}
