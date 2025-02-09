package user_service

import (
	"errors"

	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	"github.com/charmbracelet/log"
)

type UserStorage interface {
	GetById(id string) (users.User, error)
	GetByUsername(id string) (users.User, error)
	Set(users.User) error
}

type Service struct {
	storage UserStorage
	logger  *log.Logger
}

func NewService(storage UserStorage) *Service {
	return &Service{
		storage: storage,
		logger:  log.Default().WithPrefix("users"),
	}
}

// Gets looks up the user by the given id, will return a [xerrors.ErrNotExist] if nothing matched
func (s *Service) Get(id string) (users.User, error) {
	return s.storage.GetById(id)
}

// GetByUsername retrieve the given user by its name, instead of id.
// returns a [xerrors.ErrNotExist] if nothing matched
func (s *Service) GetByUsername(username string) (users.User, error) {
	return s.storage.GetByUsername(username)
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
func (s *Service) Create(user users.User) (users.User, error) {
	user.BaseItem = models.NewBaseItem()
	pwhash, err := xcrypto.HashPassword(user.PasswordHash)
	if err != nil {
		return user, err
	}
	user.PasswordHash = pwhash

	_, err = s.GetByUsername(user.Username)
	if !errors.Is(err, xerrors.ErrNotExist) {
		return user, errors.New("user already exists")
	}

	err = s.storage.Set(user)
	return user, err
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *Service) Update(user users.User) (users.User, error) {
	duser, err := s.Get(user.ID)
	if err != nil {
		return user, err
	}

	// These field may not be updated
	user.BaseItem = duser.BaseItem.Update()
	user.PasswordHash = duser.PasswordHash

	return user, s.storage.Set(user)
}

// UpdatePassword will update the password field with the new password in the database
func (s *Service) UpdatePassword(id string, newPassword []byte) error {
	user, err := s.Get(id)
	if err != nil {
		return err
	}

	user.BaseItem = user.BaseItem.Update()
	pwhash, err := xcrypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = pwhash

	return s.storage.Set(user)
}
