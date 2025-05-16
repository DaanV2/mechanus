package config

import (
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/constants"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
	"github.com/charmbracelet/log"
)

type UserConfig struct {
	ConfigDir string
	CacheDir  string
}

// GetUserConfig returns the folder for storage and config for this app
func GetUserConfig() UserConfig {
	return UserConfig{
		ConfigDir: UserConfigDir(),
		CacheDir:  UserCacheDir(),
	}
}

// UserConfigDir returns the directory the app to store its config in
func UserConfigDir() string {
	if xio.DirExists("./.config") {
		folder, err := filepath.Abs("./.config")
		if err == nil {
			return folder
		} else {
			log.Fatal("error during making ./.config an absolute folder", "error", err)
		}
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("error during checking of the user config directory", "error", err)
	}

	return filepath.Join(
		dir,
		constants.SERVICE_NAME,
	)
}

// UserCacheDir returns the directory the app to store its cache data in
func UserCacheDir() string {
	if xio.DirExists("./.cache") {
		folder, err := filepath.Abs("./.cache")
		if err == nil {
			return folder
		} else {
			log.Fatal("error during making ./.cache an absolute folder", "error", err)
		}
	}

	folder, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("error during checking of the user cache directory", "error", err)
	}

	return filepath.Join(
		folder,
		constants.SERVICE_NAME,
	)
}

// StorageFolder returns a folder where anything can be stored, expected that the given paths are relative to the storage folder
func StorageFolder(paths ...string) string {
	p := append([]string{UserCacheDir()}, paths...)
	dir := filepath.Join(p...)

	xio.MakeDirAll(dir)

	return dir
}
