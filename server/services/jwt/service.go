package jwts

import (
	"context"

	"github.com/DaanV2/mechanus/server/pkg/database"
	xrand "github.com/DaanV2/mechanus/server/pkg/rand"
	"github.com/DaanV2/mechanus/server/services/users"
)

type JTI struct {
	ID      string
	Revoked bool
}

func (j *JTI) Valid() bool {
	return !j.Revoked
}

type JWTService struct {
	jtiTable *database.Table[[]JTI]
}

func (s *JWTService) Create(ctx context.Context, user users.User) (string, error) {
	jti, err := s.GetJTI(user.ID)

	// TODO
	return "foo"
}

func (s *JWTService) Refresh(ctx context.Context, user users.User, old string) (string, error) {
	jti, err := s.GetJTI(user.ID)

	// TODO
	return "foo"
}

func (s *JWTService) Validate(ctx context.Context, token string) bool {
	// TODO
	return true
}

func (s *JWTService) GetJTI(userid string) (string, error) {
	jtis, err := s.jtiTable.Get(userid)
	if err != nil {
		return "", err
	}

	for _, jti := range jtis {
		if jti.Valid() {
			return jti.ID, nil
		}
	}

	jti := JTI{
		ID:      xrand.MustID(28),
		Revoked: false,
	}
	jtis = []JTI{jti}
	err = s.jtiTable.Set(userid, jtis)

	return jti.ID, err
}
