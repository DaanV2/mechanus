package scenes

import "github.com/DaanV2/mechanus/server/mechanus/screens"

type Manager struct {
	Screens *screens.ScreenManager
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) ScreenIDForRole(role string) (string, error) {
	panic("not implemented")
}

func (m *Manager) ScreenIDForDevice(deviceID string) (string, error) {
	panic("not implemented")
}
