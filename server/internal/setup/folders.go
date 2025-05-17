package setup

import (
	"github.com/DaanV2/mechanus/server/mechanus/paths"
)

func Folders() {
	_, _ = paths.GetAppConfigDir()
	_, _ = paths.GetStateDir()
	_, _ = paths.GetUserDataDir()
}
