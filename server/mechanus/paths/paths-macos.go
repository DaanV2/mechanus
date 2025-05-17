//go:build darwin

package paths

import (
	"os"
	"path/filepath"

	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

func getUserDataDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Application Support/AppName
	dir := filepath.Join(home, "Library", "Application Support", appName)
	xio.MakeDirAll(dir)

	return dir, nil
}

func getAppConfigDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Preferences/AppName
	dir := filepath.Join(home, "Library", "Preferences", appName)
	xio.MakeDirAll(dir)

	return dir, nil
}

func getStateDir(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ErrHomeNotFound
	}

	// macOS convention: ~/Library/Application Support/AppName/State
	dir := filepath.Join(home, "Library", "Application Support", appName, "State")
	xio.MakeDirAll(dir)

	return dir, nil
}
