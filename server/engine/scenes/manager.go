package scenes

import "github.com/DaanV2/mechanus/server/engine/screens"

// Manager manages scenes and their associated screens.
type Manager struct {
	Screens *screens.ScreenManager
}

// NewManager creates a new scene manager.
func NewManager() *Manager {
	return &Manager{}
}

// ScreenIDForRole returns the screen ID associated with a role.
func (m *Manager) ScreenIDForRole(role string) (string, error) {
	panic("not implemented")
}

// ScreenIDForDevice returns the screen ID associated with a device.
func (m *Manager) ScreenIDForDevice(deviceID string) (string, error) {
	panic("not implemented")
}
