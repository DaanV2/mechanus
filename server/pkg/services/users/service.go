package user_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/database/models"
	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	"github.com/google/uuid"
)

type Service struct {
	db     *database.DB
	logger logging.Enriched
}

func NewService(db *database.DB) *Service {
	return &Service{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("users"),
	}
}

// Gets looks up the user by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (s *Service) Get(ctx context.Context, userId string) (*models.User, error) {
	logger := s.logger.With("userId", userId).From(ctx)
	logger.Debug("Getting user by id")

	var user models.User
	tx := s.db.WithContext(ctx).First(&user, userId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

// GetByUsername retrieve the given user by its name, instead of id.
// returns a [xerrors.ErrNotExist] if nothing matched
func (s *Service) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	logger := s.logger.With("username", username).From(ctx)
	logger.Debug("Getting user by username")

	var user models.User
	tx := s.db.WithContext(ctx).First(&user, "name = ?", username)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
// It updates the user with the new password hash and sets the ID to a new UUID
func (s *Service) Create(ctx context.Context, user *models.User) error {
	logger := s.logger.With("username", user.Name).From(ctx)
	logger.Debug("Creating user")

	err := updatePassword(user)
	if err != nil {
		return err
	}

	_, err = s.GetByUsername(ctx, user.Name)
	if !database.IsNotExist(err) {
		return errors.New("user already exists")
	}

	tx := s.db.WithContext(ctx).Create(user)

	return tx.Error
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *Service) Update(ctx context.Context, user *models.User) error {
	logger := s.logger.With("userId", user.ID).From(ctx)
	logger.Debug("updating user")

	tx := s.db.WithContext(ctx).Omit("passwordhash").Updates(user)
	
	return tx.Error
}

// UpdatePassword will update the password field with the new password in the database
func (s *Service) UpdatePassword(ctx context.Context, id uuid.UUID, newPassword []byte) error {
	logger := s.logger.With("userId", id).From(ctx)
	logger.Debug("updating password")

	user := &models.User{Model: models.Model{ID: id }}
	err := updatePassword(user)
	if err != nil {
		return err
	}

	tx := s.db.WithContext(ctx).Select("passwordhash").Updates(user)
	
	return tx.Error
}

func updatePassword(user *models.User) error {
	pwhash, err := xcrypto.HashPassword(user.PasswordHash)
	if err != nil {
		return fmt.Errorf("error hashing the password: %w", err)
	}
	user.PasswordHash = pwhash
	return nil
}
