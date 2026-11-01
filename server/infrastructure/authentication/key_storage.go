package authentication

import (
	"context"
	"fmt"
	"sync"
	"time"

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

	// Guards picking / generating the signing key, so concurrent requests don't each generate a key.
	signingLock sync.Mutex
	// How long a key is used for signing before a new one is generated, 0 disables rotation.
	rotation time.Duration
	// How long tokens stay valid, a key is retired once it stopped signing longer ago than this.
	maxTokenLifetime time.Duration
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

// WithRotation enables key rotation: a new signing key is generated every rotation,
// and old keys are removed once every token they signed (valid for at most maxTokenLifetime) has expired.
func (manager *KeyManager) WithRotation(rotation, maxTokenLifetime time.Duration) *KeyManager {
	manager.rotation = rotation
	manager.maxTokenLifetime = maxTokenLifetime

	return manager
}

// AfterInitialize implements application.AfterInitialize.
func (manager *KeyManager) AfterInitialize(ctx context.Context) error {
	for k := range manager.storage.Keys(ctx) {
		key, err := manager.Get(ctx, k)
		if err != nil {
			return fmt.Errorf("error loading %s: %w", k, err)
		}

		// Keys stored before the creation time was tracked start their rotation period now
		if key.created.IsZero() {
			key.created = time.Now()
			if err := manager.save(ctx, key); err != nil {
				return fmt.Errorf("error saving %s: %w", k, err)
			}
		}
	}

	return manager.retire(ctx)
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

	key := NewKeyData(item.ID(), item.Private(), time.Now())

	return key, manager.save(ctx, key)
}

// GetSigningKey returns the newest key, generating a new one if there is none or the newest is due for rotation.
func (manager *KeyManager) GetSigningKey(ctx context.Context) (*KeyData, error) {
	manager.signingLock.Lock()
	defer manager.signingLock.Unlock()

	key := manager.newest()
	if key != nil && !manager.dueForRotation(key) {
		return key, nil
	}

	key, err := manager.New(ctx)
	if err != nil {
		return nil, err
	}

	if err := manager.retire(ctx); err != nil {
		manager.logger.From(ctx).Error("error retiring old keys", "error", err)
	}

	return key, nil
}

// newest returns the most recently created private key, ties are broken on the id so the choice is stable.
func (manager *KeyManager) newest() *KeyData {
	var result *KeyData

	for _, v := range manager.keys.Items() {
		if v == nil || v.Private() == nil {
			continue
		}

		if result == nil ||
			v.created.After(result.created) ||
			(v.created.Equal(result.created) && v.id > result.id) {
			result = v
		}
	}

	return result
}

func (manager *KeyManager) dueForRotation(key *KeyData) bool {
	return manager.rotation > 0 && time.Since(key.created) > manager.rotation
}

// retire removes the keys that stopped signing longer ago than the max token lifetime.
func (manager *KeyManager) retire(ctx context.Context) error {
	if manager.rotation <= 0 {
		return nil
	}

	newest := manager.newest()
	cutoff := manager.rotation + manager.maxTokenLifetime

	for id, key := range manager.keys.Items() {
		if key == nil || key == newest || key.created.IsZero() || time.Since(key.created) <= cutoff {
			continue
		}

		manager.logger.From(ctx).Info("retiring jwt signing key", "kid", id)
		if _, err := manager.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("error deleting key %s: %w", id, err)
		}
		manager.keys.Delete(id)
	}

	return nil
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
