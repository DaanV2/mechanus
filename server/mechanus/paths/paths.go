package paths

import (
	"errors"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/mechanus/constants"
	"github.com/charmbracelet/log"
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

// StorageFolder will return a [GetStateDir] folder appended with the subfolder
func StorageFolder(subfolder string) string {
	path, err := GetStateDir()

	log.Fatal("couldn't setup state directory", "error", err)

	return filepath.Join(path, subfolder)
}

// ConfigFolder will return a [GetAppConfigDir] folder appended with the subfolder
func ConfigFolder(subfolder string) string {
	path, err := GetAppConfigDir()

	log.Fatal("couldn't setup app config directory", "error", err)

	return filepath.Join(path, subfolder)
}

// UserFolder will return a [GetUserDataDir] folder appended with the subfolder
func UserFolder(subfolder string) string {
	path, err := GetAppConfigDir()

	log.Fatal("couldn't setup user data directory", "error", err)

	return filepath.Join(path, subfolder)
}
