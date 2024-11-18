package authenication

import (
	"errors"

	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	xsync "github.com/DaanV2/mechanus/server/pkg/extensions/sync"
	"github.com/DaanV2/mechanus/server/pkg/storage"
)

type KeyManager struct {
	storage storage.Storage[*KeyData]
	keys    *xsync.Map[string, *KeyData]
}

func NewKeyManager(storage storage.Storage[*KeyData]) (*KeyManager, error) {
	manager := &KeyManager{
		storage: storage,
		keys:    xsync.NewMap[string, *KeyData](),
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
		return item, nil
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

	for _, v := range manager.keys.Items() {
		if v == nil || v.Private() == nil {
			continue
		}

		key = v
		break
	}

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
