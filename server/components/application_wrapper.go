package components

import (
	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
)

func NewKeyManager(
	cm *application.ComponentManager,
	sp storage.StorageProvider[*authentication.KeyData],
) (*authentication.KeyManager, error) {
	manager, err := authentication.NewKeyManager(sp)
	if err == nil {
		return application.Register(cm, manager), nil
	}

	return nil, err
}
