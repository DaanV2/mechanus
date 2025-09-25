package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/repositories"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) Get(ctx context.Context, userId string) (*models.User, error) {
	return s.repo.Get(ctx, userId)
}

func (s *UserService) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.FindByUsername(ctx, username)
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
// It updates the user with the new password hash and sets the ID to a new UUID
func (s *UserService) Create(ctx context.Context, user *models.User) error {
	err := updatePassword(user)
	if err != nil {
		return err
	}

	_, err = s.repo.FindByUsername(ctx, user.Username)
	if !persistence.IsNotExist(err) {
		return ErrUserAlreadyExists
	}

	return s.repo.Create(ctx, user)
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *UserService) Update(ctx context.Context, user *models.User) error {
	return s.repo.Update(ctx, user)
}

// UpdatePassword will update the password field with the new password in the database
func (s *UserService) UpdatePassword(ctx context.Context, id string, newPassword []byte) error {
	user := &models.User{
		Model:        models.Model{ID: id},
		PasswordHash: newPassword,
	}

	err := updatePassword(user)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, user)
}

func (s *UserService) Find(ctx context.Context, queries *models.User) ([]*models.User, error) {
	return s.repo.Find(ctx, queries)
}

func updatePassword(user *models.User) error {
	pwhash, err := xcrypto.HashPassword(user.PasswordHash)
	if err != nil {
		return fmt.Errorf("error hashing the password: %w", err)
	}

	user.PasswordHash = pwhash

	return nil
}
