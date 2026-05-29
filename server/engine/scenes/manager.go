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
// TODO: Implement role-to-screen mapping (e.g. GM screen, player screen, table screen)
func (m *Manager) ScreenIDForRole(role string) (string, error) {
	panic("not implemented")
}

// ScreenIDForDevice returns the screen ID associated with a device.
// TODO: Implement device-to-screen mapping so each connected device gets assigned its screen
func (m *Manager) ScreenIDForDevice(deviceID string) (string, error) {
	panic("not implemented")
}
