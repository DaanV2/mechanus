package setup

import (
	"github.com/DaanV2/mechanus/server/pkg/config"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
)

func Folders() {
	xio.MakeDirAll(config.UserCacheDir())
	xio.MakeDirAll(config.UserConfigDir())
}