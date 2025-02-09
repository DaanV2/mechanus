package components

import (
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/google/wire"
)

var baseSet = wire.NewSet( // nolint:unused
	application.NewComponentManager,
)