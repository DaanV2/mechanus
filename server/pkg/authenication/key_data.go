package authenication

import (
	"crypto"
	"crypto/x509"
	"encoding"
	"encoding/pem"
	"errors"
)

var _ encoding.TextMarshaler = &KeyData{}
var _ encoding.TextUnmarshaler = &KeyData{}

type (
	KeyData struct {
		id  string
		key crypto.PrivateKey
	}

	publicCarrier interface {
		Public() crypto.PublicKey
	}
)

func (k *KeyData) ID() string {
	return k.id
}

func (k *KeyData) Private() crypto.PrivateKey {
	return k.key
}

func (k *KeyData) Public() crypto.PublicKey {
	if p, ok := k.key.(publicCarrier); ok {
		return p.Public()
	}

	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (k *KeyData) UnmarshalText(text []byte) error {
	data, _ := pem.Decode(text)
	if data == nil {
		return errors.New("rsa key not stored in PEM format")
	}

	id, ok := data.Headers["id"]
	if !ok {
		return errors.New("should have an id in the header of the file")
	}
	k.id = id

	akey, err := x509.ParsePKCS8PrivateKey(data.Bytes)
	if err != nil {
		return err
	}
	if akey == nil {
		return errors.New("no private key returned")
	}

	key, ok := akey.(crypto.PrivateKey)
	if !ok {
		return errors.New("no private key returned")
	}
	k.key = key

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (k *KeyData) MarshalText() (text []byte, err error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(k.key)
	if err != nil {
		return nil, err
	}

	data := &pem.Block{
		Type: "PRIVATE KEY",
		Headers: map[string]string{
			"id": k.id,
		},
		Bytes: bytes,
	}

	return pem.EncodeToMemory(data), nil
}
