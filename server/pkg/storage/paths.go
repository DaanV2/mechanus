package storage

import (
	"errors"
	"os"

	"github.com/DaanV2/mechanus/server/pkg/constants"
)

const (
	DEFAULT_PERMISSIONS = 0755
)

var (
	ErrHomeNotFound = errors.New("unable to determine user home directory")
	ErrCreateDir    = errors.New("failed to create directory")
)

// GetUserDataDir returns the directory for storing user data
func GetUserDataDir() (string, error) {
	return getUserDataDir(constants.SERVICE_NAME)
}

// GetAppConfigDir returns the directory for storing application configuration
func GetAppConfigDir() (string, error) {
	return getAppConfigDir(constants.SERVICE_NAME)
}

// GetStateDir returns the directory for storing application state
func GetStateDir() (string, error) {
	return getStateDir(constants.SERVICE_NAME)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, DEFAULT_PERMISSIONS)
}
