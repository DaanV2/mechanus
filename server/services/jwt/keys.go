package jwts

import (
	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	"github.com/charmbracelet/log"
)


type JWKS struct {
	keys []*xcrypto.RSAKey
}

func (s *JWKS) GetSigningKey() (*xcrypto.RSAKey, error) {
	for _, k := range s.keys {
		if k != nil {
			return k, nil
		}
	}

	return s.NewKey()
}

func (s *JWKS) GetKey(id string) *xcrypto.RSAKey {
	for _, k := range s.keys {
		if k.ID() == id {
			return k
		}
	}

	return nil
}

func (s *JWKS) NewKey() (*xcrypto.RSAKey, error) {
	key, err := xcrypto.GenerateRSAKeys()
	if err != nil {
		return nil, err
	}

	log.Info("generating new signing RSA key", "id", key.ID())
	s.keys = append(s.keys, key)
	return key, nil
}

