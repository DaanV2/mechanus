package screens

import (
	"context"
	"sync"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

// ScreenManager manages the screen handlers, one per screen device. Screens are keyed by device ID.
type ScreenManager struct {
	screens xsync.Map[string, *ScreenHandler]
}

// NewScreenManager creates a new screen manager.
func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		screens: xsync.Map[string, *ScreenHandler]{},
	}
}

// Get retrieves a screen handler by device ID.
func (sm *ScreenManager) Get(id string) (*ScreenHandler, bool) {
	return sm.screens.Load(id)
}

// GetOrCreate returns the screen handler of a device, creating it when the device has none yet.
// Screens are created lazily, by the first connection or the first instruction for the device.
func (sm *ScreenManager) GetOrCreate(id string) *ScreenHandler {
	if h, ok := sm.screens.Load(id); ok {
		return h
	}
	h, _ := sm.screens.LoadOrStore(id, NewScreenHandler(id))

	return h
}

// BeforeShutdown implements lifecycle.BeforeShutdown, closing all screen connections.
func (sm *ScreenManager) BeforeShutdown(_ context.Context) error {
	wg := sync.WaitGroup{}
	for _, h := range sm.screens.Items() {
		wg.Go(h.Close)
	}
	wg.Wait()

	return nil
}
