package repositories

import (
	"context"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
)

type UserRepository struct {
	db     *persistence.DB
	logger logging.Enriched
}

func NewUserRepository(db *persistence.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logging.Enriched{}.WithPrefix("users"),
	}
}

// Get looks up the user by the given id
func (repo *UserRepository) Get(ctx context.Context, userId string) (*models.User, error) {
	logger := repo.logger.With("userId", userId).From(ctx)
	logger.Debug("Getting user by id")

	var user models.User

	tx := repo.db.WithContext(ctx).First(&user, "id = ?", userId)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

// FindByUsername retrieve the given user by its name, instead of id.
func (repo *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	logger := repo.logger.With("username", username).From(ctx)
	logger.Debug("Getting user by username")

	var user models.User

	tx := repo.db.WithContext(ctx).First(&user, "name = ?", username)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
// It updates the user with the new password hash and sets the ID to a new UUID
func (repo *UserRepository) Create(ctx context.Context, user *models.User) error {
	logger := repo.logger.With("username", user.Username).From(ctx)
	logger.Debug("Creating user")

	tx := repo.db.WithContext(ctx).Create(user)

	return tx.Error
}

func (repo *UserRepository) Update(ctx context.Context, user *models.User) error {
	logger := repo.logger.With("userId", user.ID).From(ctx)
	logger.Debug("updating user")

	tx := repo.db.WithContext(ctx).
		Omit("name", "password_hash", "id").
		Updates(user)

	return tx.Error
}

// UpdatePassword will apply the given model to the database, but only on the password field
func (repo *UserRepository) UpdatePassword(ctx context.Context, user *models.User) error {
	logger := repo.logger.With("userId", user.ID).From(ctx)
	logger.Debug("updating password")

	tx := repo.db.WithContext(ctx).
		Select("password_hash").
		Updates(user)

	return tx.Error
}

func (repo *UserRepository) Find(ctx context.Context, queries *models.User) ([]*models.User, error) {
	logger := repo.logger.From(ctx)
	logger.Debug("querying users")
	var users []*models.User

	tx := repo.db.WithContext(ctx).
		Model(queries).
		Find(&users)

	return users, tx.Error
}
