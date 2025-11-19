package authentication

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/infrastructure/lifecycle"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xcrypto"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

var _ lifecycle.AfterInitialize = &KeyManager{}

// KeyManager manages cryptographic keys for JWT signing and verification.
type KeyManager struct {
	storage storage.Storage[*KeyData]
	keys    *xsync.Map[string, *KeyData]
	logger  logging.Enriched
}

// NewKeyManager creates a new key manager with the provided storage provider.
func NewKeyManager(sp storage.StorageProvider[*KeyData]) (*KeyManager, error) {
	s, err := sp.AppStorage()
	if err != nil {
		return nil, err
	}

	manager := &KeyManager{
		storage: s,
		keys:    xsync.NewMap[string, *KeyData](),
		logger:  logging.Enriched{}.WithPrefix("key_manager"),
	}

	return manager, nil
}

// AfterInitialize implements application.AfterInitialize.
func (manager *KeyManager) AfterInitialize(ctx context.Context) error {
	for k := range manager.storage.Keys(ctx) {
		_, err := manager.Get(ctx, k)
		if err != nil {
			return fmt.Errorf("error loading %s: %w", k, err)
		}
	}

	return nil
}

// Get retrieves a key by its ID, loading from storage if necessary.
func (manager *KeyManager) Get(ctx context.Context, id string) (*KeyData, error) {
	manager.logger.From(ctx).Debug("getting key: " + id)
	item, ok := manager.keys.Load(id)
	if ok {
		return item, nil
	}

	return manager.load(ctx, id)
}

// New creates and saves a new cryptographic key.
func (manager *KeyManager) New(ctx context.Context) (*KeyData, error) {
	manager.logger.From(ctx).Debug("creating new key")
	item, err := xcrypto.GenerateRSAKeys()
	if err != nil {
		return nil, err
	}

	key := &KeyData{
		id:  item.ID(),
		key: item.Private(),
	}

	return key, manager.save(ctx, key)
}

// GetSigningKey retrieves or creates a signing key for JWT tokens.
func (manager *KeyManager) GetSigningKey(ctx context.Context) (*KeyData, error) {
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
		key, err = manager.New(ctx)
	}

	return key, err
}

func (manager *KeyManager) load(ctx context.Context, id string) (*KeyData, error) {
	key, err := manager.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	manager.keys.Store(id, key)

	return key, nil
}

func (manager *KeyManager) save(ctx context.Context, item *KeyData) error {
	manager.logger.From(ctx).Debug("saving key: " + item.id)
	manager.keys.Store(item.id, item)

	return manager.storage.Set(ctx, item)
}
