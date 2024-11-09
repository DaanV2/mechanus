package users

import (
	"errors"
	"strings"

	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	"github.com/charmbracelet/log"
)

type Service struct {
	storage storage.Storage[models.User]
	logger  *log.Logger
}

func NewService(storage storage.Storage[models.User]) *Service {
	return &Service{
		storage: storage,
		logger:  log.Default().WithPrefix("users"),
	}
}

func (s *Service) Get(id string) (models.User, error) {
	return s.storage.Get(id)
}

func (s *Service) GetByUsername(username string) (models.User, error) {
	return storage.First(s.storage, func(item models.User) bool {
		return strings.EqualFold(item.Name, username)
	})
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
func (s *Service) Create(user models.User) (models.User, error) {
	user.BaseItem = models.NewBaseItem()
	pwhash, err := xcrypto.HashPassword(user.PasswordHash)
	if err != nil {
		return user, err
	}
	user.PasswordHash = pwhash

	_, err = s.storage.Get(user.ID)
	if !errors.Is(err, xerrors.ErrNotExist) {
		return user, errors.New("user already exists")
	}

	err = s.storage.Set(user.ID, user)
	return user, err
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *Service) Update(user models.User) (models.User, error) {
	duser, err := s.Get(user.ID)
	if err != nil {
		return user, err
	}

	// These field may not be updated
	user.BaseItem = duser.BaseItem.Update()
	user.PasswordHash = duser.PasswordHash

	return user, s.storage.Set(user.ID, user)
}

// UpdatePassword will update the password field with the new password in the database
func (s *Service) UpdatePassword(id string, newPassword []byte) error {
	user, err := s.storage.Get(id)
	if err != nil {
		return err
	}

	user.BaseItem = user.BaseItem.Update()
	pwhash, err := xcrypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = pwhash

	return s.storage.Set(user.ID, user)
}
