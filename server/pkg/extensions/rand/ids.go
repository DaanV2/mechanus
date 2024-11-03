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
	data := make([]byte, length*2)

	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}
