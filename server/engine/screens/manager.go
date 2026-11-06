package screens

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

// ScreenManager manages the screen handlers, one per screen device. Screens are keyed by device ID.
type ScreenManager struct {
	screens    xsync.Map[string, *ScreenHandler]
	controller atomic.Pointer[SceneController]
}

// NewScreenManager creates a new screen manager.
func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		screens: xsync.Map[string, *ScreenHandler]{},
	}
}

// SetController sets who carries out the play actions screens send. Without one they are refused.
func (sm *ScreenManager) SetController(controller SceneController) {
	sm.controller.Store(&controller)
}

func (sm *ScreenManager) currentController() SceneController {
	c := sm.controller.Load()
	if c == nil {
		return nil
	}

	return *c
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
	h := NewScreenHandler(id)
	h.controller = sm.currentController
	h, _ = sm.screens.LoadOrStore(id, h)

	return h
}

// Connections returns the connection info of every connection on every screen.
func (sm *ScreenManager) Connections() []*ConnectionInfo {
	var result []*ConnectionInfo
	for _, h := range sm.screens.Items() {
		result = append(result, h.Connections()...)
	}

	return result
}

// CloseDevice closes the connections a device opened as itself, on every screen.
func (sm *ScreenManager) CloseDevice(deviceID, reason string) {
	wg := sync.WaitGroup{}
	for _, h := range sm.screens.Items() {
		wg.Go(func() { h.CloseDevice(deviceID, reason) })
	}
	wg.Wait()
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
