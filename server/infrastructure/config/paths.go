package config

import (
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/paths"
)

func ConfigPaths() []string {
	appConfStr, _ := paths.GetAppConfigDir()
	userStr, _ := paths.GetUserDataDir()

	return []string{
		appConfStr,
		userStr,
		filepath.Join(appConfStr, ".config"),
		filepath.Join(userStr, ".config"),
		".config",
	}
}
