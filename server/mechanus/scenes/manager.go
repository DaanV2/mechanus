package scenes

import "github.com/DaanV2/mechanus/server/mechanus/screens"

type Manager struct {
	Screens *screens.Manager
}

func NewManager() *Manager {
	return &Manager{}
}
