//go:build windows

package storage

import (
	"os"
	"path/filepath"
)

func getUserDataDir(appName string) (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	dir := filepath.Join(appData, appName)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func getAppConfigDir(appName string) (string, error) {
	// On Windows, config is typically stored in the same location as user data
	return getUserDataDir(appName)
}

func getStateDir(appName string) (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", ErrHomeNotFound
		}
		localAppData = filepath.Join(home, "AppData", "Local")
	}

	dir := filepath.Join(localAppData, appName)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}
