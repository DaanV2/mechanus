package storage

import (
	"github.com/DaanV2/mechanus/server/pkg/paths"
)

func SetupFolders() {
	_, _ = paths.GetAppConfigDir()
	_, _ = paths.GetStateDir()
	_, _ = paths.GetUserDataDir()
}
