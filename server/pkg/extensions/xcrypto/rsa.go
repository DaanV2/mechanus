package xcrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xrand"
)

type RSAKey struct {
	id  string
	key *rsa.PrivateKey
}

func (k *RSAKey) ID() string {
	return k.id
}

func (k *RSAKey) Private() *rsa.PrivateKey {
	return k.key
}

func (k *RSAKey) Public() crypto.PublicKey {
	return k.key.Public()
}

func GenerateRSAKeys() (*RSAKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &RSAKey{
		id:  xrand.MustID(28),
		key: privateKey,
	}, nil
}
