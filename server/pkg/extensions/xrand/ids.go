package xrand

import (
	"crypto/rand"
	"encoding/hex"
)

func MustID(length int) string {
	id, err := ID(length)
	if err != nil {
		panic(err)
	}

	return id
}

func ID(length int) (string, error) {
	l := length / 2
	if length&1 == 1 {
		l += 1
	}

	data := make([]byte, l)

	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(data)[:length], nil
}
