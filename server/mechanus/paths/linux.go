//go:build linux

package paths

import (
	"os"
	"path/filepath"

	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

func getUserDataDir(appName string) (string, error) {
	// Follow XDG Base Directory Specification
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", ErrHomeNotFound
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	dir := filepath.Join(dataHome, appName)
	xio.MakeDirAll(dir)

	return dir, nil
}

func getAppConfigDir(appName string) (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", ErrHomeNotFound
		}
		configHome = filepath.Join(home, ".config")
	}

	dir := filepath.Join(configHome, appName)
	xio.MakeDirAll(dir)

	return dir, nil
}

func getStateDir(appName string) (string, error) {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", ErrHomeNotFound
		}
		stateHome = filepath.Join(home, ".local", "state")
	}

	dir := filepath.Join(stateHome, appName)
	xio.MakeDirAll(dir)

	return dir, nil
}
