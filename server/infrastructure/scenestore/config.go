package scenestore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
)

var (
	// ErrInvalidID is returned for IDs that cannot safely be used as a storage key / file name.
	ErrInvalidID = errors.New("invalid id")
	// ErrRoomNotFound is returned when the configured room does not exist.
	ErrRoomNotFound = errors.New("room not found")
)

// ConfigStore reads authored scene and room configuration. It never writes: configuration is
// prepared by the GM, the server only reads it.
type ConfigStore struct {
	scenes storage.Storage[*scenes.Config]
	rooms  storage.Storage[*devices.Room]
	roomID string
}

// NewConfigStore creates a config store on top of the given storages. roomID selects the room
// returned by [ConfigStore.CurrentRoom].
func NewConfigStore(sceneStorage storage.Storage[*scenes.Config], roomStorage storage.Storage[*devices.Room], roomID string) *ConfigStore {
	return &ConfigStore{scenes: sceneStorage, rooms: roomStorage, roomID: roomID}
}

// NewFileConfigStore creates a config store reading from the app config folder.
func NewFileConfigStore(roomID string) (*ConfigStore, error) {
	sceneStorage, err := storage.NamedFileStorage[*scenes.Config]("scenes").AppStorage()
	if err != nil {
		return nil, fmt.Errorf("scene config storage: %w", err)
	}
	roomStorage, err := storage.NamedFileStorage[*devices.Room]("rooms").AppStorage()
	if err != nil {
		return nil, fmt.Errorf("room config storage: %w", err)
	}

	return NewConfigStore(sceneStorage, roomStorage, roomID), nil
}

// GetSceneConfig implements application.SceneConfigStore.
func (c *ConfigStore) GetSceneConfig(ctx context.Context, sceneID string) (*scenes.Config, error) {
	if err := validateID(sceneID); err != nil {
		return nil, err
	}

	cfg, err := c.scenes.Get(ctx, sceneID)
	if isNotFound(err) {
		return nil, fmt.Errorf("%q: %w", sceneID, scenes.ErrSceneNotFound)
	}
	if err != nil {
		return nil, err
	}
	if cfg.ID != sceneID {
		return nil, fmt.Errorf("scene config %q declares id %q: %w", sceneID, cfg.ID, ErrInvalidID)
	}

	return cfg, nil
}

// CurrentRoom implements application.RoomConfigStore.
func (c *ConfigStore) CurrentRoom(ctx context.Context) (*devices.Room, error) {
	if err := validateID(c.roomID); err != nil {
		return nil, err
	}

	room, err := c.rooms.Get(ctx, c.roomID)
	if isNotFound(err) {
		return nil, fmt.Errorf("%q: %w", c.roomID, ErrRoomNotFound)
	}

	return room, err
}

// validateID only allows IDs that are safe as file names, as file storage uses them as such.
func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("empty: %w", ErrInvalidID)
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return fmt.Errorf("%q: %w", id, ErrInvalidID)
		}
	}

	return nil
}

func isNotFound(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || persistence.IsNotExist(err)
}
