package components

import (
	"github.com/DaanV2/mechanus/server/pkg/authenication"
	"github.com/DaanV2/mechanus/server/pkg/models/users"
	user_service "github.com/DaanV2/mechanus/server/pkg/services/users"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	cache_storage "github.com/DaanV2/mechanus/server/pkg/storage/cache"
	file_storage "github.com/DaanV2/mechanus/server/pkg/storage/files"
	memory_storage "github.com/DaanV2/mechanus/server/pkg/storage/memory"
	user_storage "github.com/DaanV2/mechanus/server/pkg/storage/user"
	"github.com/google/wire"
)

type Storage struct {
	Users    storage.Storage[users.User]
	JTIs     storage.Storage[[]authenication.JTI]
	AuthKeys storage.Storage[*authenication.KeyData]
}

var fileStorage = wire.NewSet(
	FileStorage,
	user_storage.NewStorage,

	wire.FieldsOf(new(*Storage), "Users", "JTIs", "AuthKeys"),
	wire.Bind(new(user_service.UserStorage), new(*user_storage.Storage)),
)

var memoryStorage = wire.NewSet(
	MemoryStorage,
	user_storage.NewStorage,

	wire.FieldsOf(new(*Storage), "Users", "JTIs", "AuthKeys"),
	wire.Bind(new(user_service.UserStorage), new(*user_storage.Storage)),
)

func MemoryStorage() *Storage {
	return &Storage{
		Users:    memory_storage.NewStorage[users.User](),
		JTIs:     memory_storage.NewStorage[[]authenication.JTI](),
		AuthKeys: memory_storage.NewStorage[*authenication.KeyData](),
	}
}

func FileStorage(folder string) *Storage {
	return &Storage{
		Users:    cache_storage.NewCache(file_storage.NewStorage[users.User](folder)),
		JTIs:     cache_storage.NewCache(file_storage.NewStorage[[]authenication.JTI](folder)),
		AuthKeys: cache_storage.NewCache(file_storage.NewStorage[*authenication.KeyData](folder)),
	}
}
