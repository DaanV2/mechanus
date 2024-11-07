package authenication

import (
	"errors"

	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

type JTI struct {
	ID      string
	Revoked bool
}

func (j *JTI) Valid() bool {
	return len(j.ID) > 0 && !j.Revoked
}

type JTIService struct {
	storage storage.Storage[[]JTI]
}

func (s *JTIService) GetOrCreate(userId string) (string, error) {
	if userId == "" {
		return "", errors.New("userId is empty")
	}

	jtis, err := s.storage.Get(userId)
	if err != nil {
		return "", err
	}

	// Find valid JTI
	for _, jti := range jtis {
		if jti.Valid() {
			return jti.ID, nil
		}
	}

	jti := JTI{
		ID:      xrand.MustID(28),
		Revoked: false,
	}

	err = s.storage.Set(userId, []JTI{jti})

	return jti.ID, err
}

func (s *JTIService) Find(userId string, jti string) (JTI, error) {
	jtis, err := s.storage.Get(userId)
	if err != nil {
		return JTI{}, err
	}

	for _, j := range jtis {
		if j.ID == jti {
			return j, nil
		}
	}

	return JTI{}, storage.ErrNotExist
}
