package screens

import xsync "github.com/DaanV2/mechanus/server/pkg/extensions/sync"

type ScreenManager struct {
	screens xsync.Map[string, *ScreenHandler]
}

func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		screens: xsync.Map[string, *ScreenHandler]{},
	}
}

func (sm *ScreenManager) Get(id string) (*ScreenHandler, bool) {
	return sm.screens.Load(id)
}
