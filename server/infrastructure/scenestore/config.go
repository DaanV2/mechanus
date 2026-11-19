package scenestore

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/configfile"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/paths"
)

// RoomID is the ID of the room config file. A running instance serves exactly one game table, so there
// is exactly one room: <app config dir>/rooms/room.yaml (or .yml / .json).
const RoomID = "room"

// ConfigLogger is the logger prefix of the config store.
const ConfigLogger logging.Prefixed = "scene-config"

var (
	// ErrInvalidID is returned for IDs that cannot safely be used as a storage key / file name.
	ErrInvalidID = errors.New("invalid id")
	// ErrDuplicateID is reported when more than one file defines the same scene or room.
	ErrDuplicateID = errors.New("more than one file for the same id")
)

// SceneMigrations upgrades older scene config files, see [configfile.Migrations].
var SceneMigrations = configfile.Migrations{}

// RoomMigrations upgrades older room config files, see [configfile.Migrations].
var RoomMigrations = configfile.Migrations{}

// FileError is a config file that could not be loaded. The scene or room it defines is skipped.
type FileError struct {
	File    string // Path relative to the app config dir, e.g. scenes/cave.yaml
	Message string
}

// ConfigStore reads authored scene and room configuration from config files (YAML or JSON, see
// [configfile]): <dir>/scenes/<scene id>.yaml and <dir>/rooms/<room id>.yaml. The server never writes them.
//
// Files are read by [ConfigStore.Load] when the server starts; edits need a restart. A file that fails
// to load is skipped and reported by [ConfigStore.Errors].
type ConfigStore struct {
	dir    string // The app config dir; scene and room folders are below it
	roomID string

	lock   sync.RWMutex
	scenes map[string]*scenes.Config
	room   *devices.Room
	errors map[string]error // Relative file path -> error
}

const (
	scenesFolder = "scenes"
	roomsFolder  = "rooms"
)

// NewConfigStore creates a config store reading from the scenes and rooms folders below dir. roomID
// selects the room returned by [ConfigStore.CurrentRoom]. Call [ConfigStore.Load] before use.
func NewConfigStore(dir, roomID string) *ConfigStore {
	return &ConfigStore{
		dir:    dir,
		roomID: roomID,
		scenes: map[string]*scenes.Config{},
		errors: map[string]error{},
	}
}

// NewFileConfigStore creates a config store reading from the app config folder and loads it.
func NewFileConfigStore(ctx context.Context, roomID string) (*ConfigStore, error) {
	dir, err := paths.GetAppConfigDir()
	if err != nil {
		return nil, fmt.Errorf("app config dir: %w", err)
	}
	for _, folder := range []string{scenesFolder, roomsFolder} {
		if err := os.MkdirAll(filepath.Join(dir, folder), 0o750); err != nil {
			return nil, fmt.Errorf("creating %s folder: %w", folder, err)
		}
	}

	store := NewConfigStore(dir, roomID)
	if err := validateID(roomID); err != nil {
		return nil, fmt.Errorf("room: %w", err)
	}
	store.Load(ctx)

	return store, nil
}

// Load reads every scene file and the room file, replacing what was loaded before. Files that cannot
// be loaded are skipped and reported by [ConfigStore.Errors].
func (c *ConfigStore) Load(ctx context.Context) {
	loaded := map[string]*scenes.Config{}
	errs := map[string]error{}

	files := c.sceneFiles(ctx)
	ids := make([]string, 0, len(files))
	for id := range files {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		cfg, file, err := loadScene(files[id], id)
		switch {
		case err != nil:
			errs[file] = err
			ConfigLogger.From(ctx).Warn("scene config could not be loaded, skipping it", "file", file, "error", err)
		case cfg != nil:
			loaded[id] = cfg
		}
	}

	room, file, err := c.loadRoom(c.filesFor(ctx, roomsFolder, c.roomID))
	if err != nil {
		errs[file] = err
		ConfigLogger.From(ctx).Warn("room config could not be loaded, skipping it", "file", file, "error", err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.scenes, c.room, c.errors = loaded, room, errs
}

// Errors returns the config files that could not be loaded, sorted by file.
func (c *ConfigStore) Errors() []FileError {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]FileError, 0, len(c.errors))
	for file, err := range c.errors {
		result = append(result, FileError{File: file, Message: err.Error()})
	}
	slices.SortFunc(result, func(a, b FileError) int { return cmp.Compare(a.File, b.File) })

	return result
}

// GetSceneConfig implements application.SceneConfigStore.
func (c *ConfigStore) GetSceneConfig(_ context.Context, sceneID string) (*scenes.Config, error) {
	if err := validateID(sceneID); err != nil {
		return nil, err
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	cfg, ok := c.scenes[sceneID]
	if !ok {
		return nil, fmt.Errorf("%q: %w", sceneID, scenes.ErrSceneNotFound)
	}

	return cfg, nil
}

// ListSceneConfigs implements application.SceneConfigStore. Scene files that cannot be loaded are not
// listed; see [ConfigStore.Errors].
func (c *ConfigStore) ListSceneConfigs(_ context.Context) ([]*scenes.Config, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]*scenes.Config, 0, len(c.scenes))
	for _, cfg := range c.scenes {
		result = append(result, cfg)
	}
	slices.SortFunc(result, func(a, b *scenes.Config) int { return cmp.Compare(a.ID, b.ID) })

	return result, nil
}

// CurrentRoom implements application.RoomConfigStore. A room that is not set up yet has no devices:
// scenes can still be activated, nothing is instructed.
func (c *ConfigStore) CurrentRoom(_ context.Context) (*devices.Room, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.room == nil {
		return devices.NewRoom(c.roomID, ""), nil
	}

	return c.room, nil
}

// loadScene decodes the file of a scene.
func loadScene(files []configFile, id string) (*scenes.Config, string, error) {
	return loadOne(files, SceneMigrations, func(cfg *scenes.Config) error {
		if err := validateID(id); err != nil {
			return err
		}
		if cfg.ID == "" {
			cfg.ID = id
		}
		if cfg.ID != id {
			return fmt.Errorf("declares id %q, expected the file name %q: %w", cfg.ID, id, ErrInvalidID)
		}

		return nil
	})
}

// loadRoom decodes the room file.
func (c *ConfigStore) loadRoom(files []configFile) (*devices.Room, string, error) {
	return loadOne(files, RoomMigrations, func(room *devices.Room) error {
		if room.ID == "" {
			room.ID = c.roomID
		}
		if room.ID != c.roomID {
			return fmt.Errorf("declares id %q, expected the file name %q: %w", room.ID, c.roomID, ErrInvalidID)
		}
		if room.Devices == nil {
			room.Devices = map[string]*devices.Device{}
		}
		if room.Bindings == nil {
			room.Bindings = map[string][]string{}
		}

		return nil
	})
}

// loadOne decodes the single config file of an ID. No files means the ID is not configured (nil, no
// error). file is the file the error is about, relative to the app config dir.
func loadOne[T any](files []configFile, migrations configfile.Migrations, validate func(*T) error) (result *T, file string, err error) {
	switch len(files) {
	case 0:
		return nil, "", nil
	case 1:
	default:
		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, f.rel)
		}

		return nil, files[0].rel, fmt.Errorf("%w: %s, remove all but one", ErrDuplicateID, strings.Join(names, ", "))
	}

	f := files[0]
	if !configfile.IsConfigFile(f.rel) {
		return nil, f.rel, fmt.Errorf("%w: %q, rename the file to .yaml or .json and add \"version: 1\"", configfile.ErrUnsupportedFormat, filepath.Ext(f.rel))
	}
	//nolint:gosec // the path is built from the config dir and a directory listing
	data, err := os.ReadFile(f.abs)
	if errors.Is(err, fs.ErrNotExist) {
		// Removed between listing and reading
		return nil, "", nil
	}
	if err != nil {
		return nil, f.rel, err
	}

	result = new(T)
	if err := configfile.Decode(f.rel, data, migrations, result); err != nil {
		return nil, f.rel, err
	}
	if err := validate(result); err != nil {
		return nil, f.rel, err
	}

	return result, f.rel, nil
}

type configFile struct {
	abs string
	rel string // Relative to the app config dir, with forward slashes for display
}

// sceneFiles returns the config file candidates in the scenes folder, by scene ID.
func (c *ConfigStore) sceneFiles(ctx context.Context) map[string][]configFile {
	result := map[string][]configFile{}
	for _, f := range c.listFolder(ctx, scenesFolder) {
		id := configfile.ID(f.rel)
		result[id] = append(result[id], f)
	}

	return result
}

// filesFor returns the config file candidates of one ID in a folder.
func (c *ConfigStore) filesFor(ctx context.Context, folder, id string) []configFile {
	var result []configFile
	for _, f := range c.listFolder(ctx, folder) {
		if configfile.ID(f.rel) == id {
			result = append(result, f)
		}
	}

	return result
}

// listFolder lists the files of a config folder that are, or look like they were meant to be, config
// files: supported formats and the legacy .dat files (which are reported). Sub folders hold scene
// assets and are skipped.
func (c *ConfigStore) listFolder(ctx context.Context, folder string) []configFile {
	dir := filepath.Join(c.dir, folder)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			ConfigLogger.From(ctx).Warn("could not list config folder", "folder", dir, "error", err)
		}

		return nil
	}

	var result []configFile
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !isCandidate(name) {
			continue
		}
		result = append(result, configFile{
			abs: filepath.Join(dir, name),
			rel: folder + "/" + name,
		})
	}

	return result
}

// isCandidate reports whether a file in a config folder is meant as a config file.
func isCandidate(name string) bool {
	if strings.HasPrefix(name, ".") {
		return false // Hidden and editor swap files
	}

	return configfile.IsConfigFile(name) || strings.EqualFold(filepath.Ext(name), ".dat")
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
