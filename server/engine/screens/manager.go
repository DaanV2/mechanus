package screens

import (
	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

// ScreenManager manages multiple screen handlers.
type ScreenManager struct {
	screens xsync.Map[string, *ScreenHandler]
}

// NewScreenManager creates a new screen manager.
func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		screens: xsync.Map[string, *ScreenHandler]{},
	}
}

// Get retrieves a screen handler by ID.
func (sm *ScreenManager) Get(id string) (*ScreenHandler, bool) {
	return sm.screens.Load(id)
}
