package user_storage

import (
	"errors"

	xerrors "github.com/DaanV2/mechanus/server/pkg/extensions/errors"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

type Storage struct {
	base storage.Storage[users.User]
}

func NewStorage(base storage.Storage[users.User]) *Storage {
	return &Storage{base}
}

func (s *Storage) GetById(id string) (users.User, error) {
	return s.base.Get(id)
}

func (s *Storage) GetByUsername(id string) (users.User, error) {
	for id := range s.base.Ids() {
		u, err := s.GetById(id)
		if err != nil {
			continue
		}

		if u.Username == id {
			return u, nil
		}
	}

	return users.User{}, xerrors.ErrNotExist
}

func (s *Storage) Set(u users.User) error {
	if u.ID == "" {
		return errors.New("empty id")
	}

	return s.base.Set(u.ID, u)
}