//go:build darwin

package storage

import (
	"os"
	"path/filepath"
)

func getUserDataDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Application Support/AppName
	dir := filepath.Join(home, "Library", "Application Support", appName)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func getAppConfigDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Preferences/AppName
	dir := filepath.Join(home, "Library", "Preferences", appName)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func getStateDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Application Support/AppName/State
	dir := filepath.Join(home, "Library", "Application Support", appName, "State")
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}
