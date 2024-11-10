package authenication

import (
	"errors"
	"sync"

	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

type KeyManager struct {
	storage storage.Storage[*KeyData]
	keys    sync.Map
}

func NewKeyManager(storage storage.Storage[*KeyData]) (*KeyManager, error) {
	manager := &KeyManager{
		storage: storage,
		keys:    sync.Map{},
	}

	var err error
	for id := range storage.Ids() {
		_, cerr := manager.load(id)
		err = errors.Join(err, cerr)
	}

	return manager, nil
}

func (manager *KeyManager) Get(id string) (*KeyData, error) {
	item, ok := manager.keys.Load(id)
	if ok {
		key, ok := item.(*KeyData)
		if ok {
			return key, nil
		}
	}

	return manager.load(id)
}

func (manager *KeyManager) New() (*KeyData, error) {
	item, err := xcrypto.GenerateRSAKeys()
	if err != nil {
		return nil, err
	}

	key := &KeyData{
		id:  item.ID(),
		key: item.Private(),
	}
	return key, manager.save(key)
}

// GetSigningKey
func (manager *KeyManager) GetSigningKey() (*KeyData, error) {
	var (
		key *KeyData
		err error
	)

	manager.keys.Range(func(id, value any) bool {
		k, ok := value.(*KeyData)
		if !ok || k.Private() == nil {
			return true
		}

		key = k
		return false
	})

	if key == nil {
		key, err = manager.New()
	}

	return key, err
}

func (manager *KeyManager) load(id string) (*KeyData, error) {
	key, err := manager.storage.Get(id)
	if err != nil {
		return nil, err
	}

	manager.keys.Store(id, key)
	return key, nil
}

func (manager *KeyManager) save(item *KeyData) error {
	manager.keys.Store(item.id, item)
	return manager.storage.Set(item.id, item)
}
