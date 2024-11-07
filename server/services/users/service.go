package users

import (
	"errors"
	"strings"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	TABLE_USER database.TableName = "users"
)

type Service struct {
	storage storage.Storage[models.Campaign]
	logger    *log.Logger
}

func NewService(db *database.Database) *Service {
	return &Service{
		db:        db,
		userTable: database.GetTable[models.User](db, TABLE_USER),
		logger:    log.Default().WithPrefix("users"),
	}
}

func (s *Service) Get(id string) (models.User, error) {
	return s.userTable.Get(id)
}

func (s *Service) GetByUsername(username string) (models.User, error) {
	return s.userTable.First(func(item models.User) bool {
		return strings.EqualFold(item.Name, username)
	})
}

// Create makes a new entry in the database, assumes the password is set in the PasswordHash field as plain bytes, will hash that field first
func (s *Service) Create(user models.User) (models.User, error) {
	user.BaseItem = models.NewBaseItem()
	err := HashPassword(&user)
	if err != nil {
		return user, err
	}

	_, err = s.userTable.Get(user.ID)
	if !errors.Is(err, database.ErrNotFound) {
		return user, database.ErrAlreadyExists
	}

	err = s.userTable.Set(user.ID, user)
	return user, err
}

// Update will take the new information in the user and update the database entry. Note, this does not update the password or the ID
func (s *Service) Update(user User) (User, error) {
	duser, err := s.Get(user.ID)
	if err != nil {
		return user, err
	}

	// These field may not be updated
	user.BaseItem = duser.BaseItem.Update()
	user.PasswordHash = duser.PasswordHash

	return user, s.userTable.Set(user.ID, user)
}

// UpdatePassword will update the password field with the new password in the database
func (s *Service) UpdatePassword(id string, newPassword []byte) error {
	user, err := s.userTable.Get(id)
	if err != nil {
		return err
	}

	user.PasswordHash = newPassword
	user.BaseItem = user.BaseItem.Update()
	err = HashPassword(&user)
	if err != nil {
		return err
	}

	return s.userTable.Set(user.ID, user)
}

func HashPassword(user *User) error {
	pwhash, err := bcrypt.GenerateFromPassword(user.PasswordHash, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = pwhash

	return nil
}

func ComparePassword(hashedPassword, password []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}
