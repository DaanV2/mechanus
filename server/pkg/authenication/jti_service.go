package authenication

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/database"
)

type JTI struct {
	ID      string
	UserId  string
	Revoked bool
}

func (j *JTI) Valid() bool {
	return len(j.ID) > 0 && !j.Revoked
}

type JTIService struct {
	db *database.DB
	logger logging.Enriched
}

func NewJTIService(db *database.DB) *JTIService {
	return &JTIService{
		db: db,
		logger: logging.Enriched{}.WithPrefix("jti_service"),
	}
}


func (s *JTIService) GetOrCreate(userId string) (string, error) {
	if userId == "" {
		return "", errors.New("userId is empty")
	}

	return jti.ID, err
}


func (s *JTIService) Get(userId string) (string, error) {
	if userId == "" {
		return "", errors.New("userId is empty")
	}

	return jti.ID, err
}


func (s *JTIService) Create(ctx context.Context, userId string) (*JTI, error) {
	if userId == "" {
		return "", errors.New("userId is empty")
	}

	result := JTI{
		
	}
	tx := s.db.WithContext(ctx).Create(result)

	return jti.ID, err
}

func (s *JTIService) Revoke() {

}

func (s *JTIService) Find(ctx context.Context, jti string) (*JTI, error) {
	logger := s.logger.From(ctx).With("jti", jti)
	logger.Debug("getting jti")

	var result JTI
	tx := s.db.WithContext(ctx).First(&result, "id = ? ", jti)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &result, nil
}
