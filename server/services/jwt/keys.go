package jwts

import (
	"crypto/rand"
	"crypto/rsa"

	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"
	"github.com/charmbracelet/log"
)

type Key struct {
	id  string
	key *rsa.PrivateKey
}

type JWKS struct {
	keys []*Key
}

func (s *JWKS) GetSigningKey() (*Key, error) {
	for _, k := range s.keys {
		if k != nil {
			return k, nil
		}
	}

	return s.NewKey()
}

func (s *JWKS) GetKey(id string) *Key {
	for _, k := range s.keys {
		if k.id == id {
			return k
		}
	}

	return nil
}

func (s *JWKS) NewKey() (*Key, error) {
	k, err := GenerateRSAKeys()
	if err != nil {
		return nil, err
	}

	log.Info("generating new signing RSA key", "id", k.id)
	s.keys = append(s.keys, k)
	return k, nil
}

