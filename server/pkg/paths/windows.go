//go:build windows

package paths

import (
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xio"
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
	xio.MakeDirAll(dir)

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
	xio.MakeDirAll(dir)

	return dir, nil
}
