package storage

import (
	"github.com/DaanV2/mechanus/server/pkg/paths"
)

// SetupFolders ensures that required application directories exist.
func SetupFolders() {
	_, _ = paths.GetAppConfigDir()
	_, _ = paths.GetStateDir()
	_, _ = paths.GetUserDataDir()
}
