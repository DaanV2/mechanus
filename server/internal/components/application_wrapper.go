package components

import (
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

func NewKeyManager(
	cm *application.ComponentManager,
	sp storage.StorageProvider[*authenication.KeyData],
) (*authenication.KeyManager, error) {
	manager, err := authenication.NewKeyManager(sp)
	if err == nil {
		return application.Register(cm, manager), nil
	}
	return nil, err
}
