package screens

import "github.com/DaanV2/mechanus/server/mechanus/paths"

type Manager struct {
	folder string
}

func NewManager() (*Manager, error) {
	f, err := paths.GetAppConfigDir()
	if err != nil {
		return nil, err
	}

	return &Manager{
		folder: f,
	}, nil
}
