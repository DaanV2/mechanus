package xcrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xrand"
)

// RSAKey represents an RSA private key with an associated ID.
type RSAKey struct {
	id  string
	key *rsa.PrivateKey
}

// ID returns the unique identifier for this RSA key.
func (k *RSAKey) ID() string {
	return k.id
}

// Private returns the RSA private key.
func (k *RSAKey) Private() *rsa.PrivateKey {
	return k.key
}

// Public returns the RSA public key.
func (k *RSAKey) Public() crypto.PublicKey {
	return k.key.Public()
}

// GenerateRSAKeys generates a new 2048-bit RSA key pair with a random ID.
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
