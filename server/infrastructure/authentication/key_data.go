package authentication

import (
	"crypto"
	"crypto/x509"
	"encoding"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

var _ encoding.TextMarshaler = &KeyData{}
var _ encoding.TextUnmarshaler = &KeyData{}

const (
	keyHeaderID      = "id"
	keyHeaderCreated = "created"
)

type (
	// KeyData holds a cryptographic key with an identifier.
	KeyData struct {
		id  string
		key crypto.PrivateKey
		// When the key was generated, zero for keys stored before this was tracked.
		created time.Time
	}

	publicCarrier interface {
		Public() crypto.PublicKey
	}
)

// NewKeyData wraps a private key with the given identifier and creation time.
func NewKeyData(id string, key crypto.PrivateKey, created time.Time) *KeyData {
	return &KeyData{
		id:      id,
		key:     key,
		created: created,
	}
}

// GetID returns the key's identifier.
func (k *KeyData) GetID() string {
	return k.id
}

// Created returns when the key was generated, zero if unknown.
func (k *KeyData) Created() time.Time {
	return k.created
}

// Private returns the private key.
func (k *KeyData) Private() crypto.PrivateKey {
	return k.key
}

// Public returns the public key corresponding to the private key.
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

	id, ok := data.Headers[keyHeaderID]
	if !ok {
		return errors.New("should have an id in the header of the file")
	}

	k.id = id

	k.created = time.Time{}
	if created, ok := data.Headers[keyHeaderCreated]; ok {
		t, err := time.Parse(time.RFC3339, created)
		if err != nil {
			return fmt.Errorf("invalid created header: %w", err)
		}
		k.created = t
	}

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

	headers := map[string]string{
		keyHeaderID: k.id,
	}
	if !k.created.IsZero() {
		headers[keyHeaderCreated] = k.created.UTC().Format(time.RFC3339)
	}

	data := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: headers,
		Bytes:   bytes,
	}

	return pem.EncodeToMemory(data), nil
}
