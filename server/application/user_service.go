package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService struct {
	db     *persistence.DB
	logger logging.Enriched
}

func NewUserService(db *persistence.DB) *UserService {
	return &UserService{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("users"),
	}
}

// Gets looks up the user by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (s *UserService) Get(ctx context.Context, userId string) (*models.User, error) {
	logger := s.logger.With("userId", userId).From(ctx)
	logger.Debug("Getting user by id")

	var user models.User

	tx := s.db.WithContext(ctx).First(&user, "id = ?", userId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

// GetByUsername retrieve the given user by its name, instead of id.
// returns a [xerrors.ErrNotExist] if nothing matched
func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
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
func (s *UserService) Create(ctx context.Context, user *models.User) error {
	logger := s.logger.With("username", user.Name).From(ctx)
	logger.Debug("Creating user")

	err := updatePassword(user)
	if err != nil {
		return err
	}

	_, err = s.GetByUsername(ctx, user.Name)
	if !persistence.IsNotExist(err) {
		return ErrUserAlreadyExists
	}

	tx := s.db.WithContext(ctx).Create(user)

	return tx.Error
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *UserService) Update(ctx context.Context, user *models.User) error {
	logger := s.logger.With("userId", user.ID).From(ctx)
	logger.Debug("updating user")

	// TODO ensure that the name cannot be updated, or stays unique
	tx := s.db.WithContext(ctx).Omit("password_hash", "id").Updates(user)

	return tx.Error
}

// UpdatePassword will update the password field with the new password in the database
func (s *UserService) UpdatePassword(ctx context.Context, id string, newPassword []byte) error {
	logger := s.logger.With("userId", id).From(ctx)
	logger.Debug("updating password")

	user := &models.User{Model: models.Model{ID: id}}

	err := updatePassword(user)
	if err != nil {
		return err
	}

	tx := s.db.WithContext(ctx).Select("password_hash").Updates(user)

	return tx.Error
}

func (s *UserService) Find(ctx context.Context, queries *models.User) ([]*models.User, error) {
	var users []*models.User

	tx := s.db.WithContext(ctx).Model(queries).Find(&users)

	return users, tx.Error
}

func updatePassword(user *models.User) error {
	pwhash, err := xcrypto.HashPassword(user.PasswordHash)
	if err != nil {
		return fmt.Errorf("error hashing the password: %w", err)
	}

	user.PasswordHash = pwhash

	return nil
}
