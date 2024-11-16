//go:build wireinject
// +build wireinject

package components

import (
	"github.com/DaanV2/mechanus/server/pkg/config"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	file_storage "github.com/DaanV2/mechanus/server/pkg/storage/files"
	"github.com/DaanV2/mechanus/server/services/users"
	"github.com/google/wire"
)

var (
	configSet = wire.NewSet(
		config.GetUserConfig,
	)
	storageSet = wire.NewSet(
		newUserStorage,
	)
)

func NewUserService() *users.Service {
	wire.Build(configSet, storageSet, users.NewService)

	return &users.Service{}
}

// Storage
func newUserStorage(conf config.UserConfig) storage.Storage[models.User] {
	return file_storage.NewStorage[models.User](conf.CacheDir)
}
