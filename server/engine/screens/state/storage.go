package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xio"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xos"
	"github.com/DaanV2/mechanus/server/pkg/paths"
)

var (
	ErrScreenNotFound = errors.New("screen not found")
)

type ScreenStorage struct {
	// Each screen has its own storage subfolder.
	folder string
}

func NewScreenStorage() *ScreenStorage {
	return &ScreenStorage{folder: paths.StorageFolder("screens")}
}

// DeleteScreen deletes a screen and its metadata.
func (s *ScreenStorage) DeleteScreen(screenID string) error {
	if !xio.DirExists(filepath.Join(s.folder, screenID)) {
		return ErrScreenNotFound
	}

	if err := os.RemoveAll(filepath.Join(s.folder, screenID)); err != nil {
		return fmt.Errorf("failed to delete screen folder: %w", err)
	}

	return nil
}

// ListScreens lists all screens stored in the storage folder.
func (s *ScreenStorage) ListScreens() ([]string, error) {
	files, err := os.ReadDir(s.folder)
	if err != nil {
		return nil, fmt.Errorf("failed to read screens folder: %w", err)
	}

	var screens []string
	for _, file := range files {
		if file.IsDir() {
			screens = append(screens, file.Name())
		}
	}

	return screens, nil
}

// GetScreenMetadata retrieves the metadata for a screen by its ID.
// returns [ErrScreenNotFound] if the screen does not exist.
func (s *ScreenStorage) GetScreenMetadata(screenID string) (*ScreenMetadata, error) {
	f := filepath.Clean(filepath.Join(s.folder, screenID))
	filePath := filepath.Join(f, "metadata.json")
	if !xio.DirExists(f) {
		return nil, ErrScreenNotFound
	}
	data, err := os.ReadFile(filePath) //nolint:gosec // G304 ignored cause we clean the path
	if err != nil {
		return nil, fmt.Errorf("failed to read screen metadata file: %w", err)
	}
	var metadata ScreenMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal screen metadata: %w", err)
	}

	return &metadata, nil
}

// SaveScreenMetadata saves the metadata for a screen.
func (s *ScreenStorage) SaveScreenMetadata(screenID string, metadata *ScreenMetadata) error {
	f := filepath.Clean(filepath.Join(s.folder, screenID))
	filePath := filepath.Join(f, "metadata.json")
	xio.MakeDirAll(f)

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal screen metadata: %w", err)
	}
	if err := xos.WriteFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write screen metadata to file: %w", err)
	}

	return nil
}

// GetScreenState retrieves the state for a screen by its ID.
func (s *ScreenStorage) GetScreenState(screenID string) (*ScreenState, error) {
	f := filepath.Clean(filepath.Join(s.folder, screenID))
	filePath := filepath.Join(f, "state.json")
	if !xio.DirExists(f) {
		return nil, ErrScreenNotFound
	}

	data, err := os.ReadFile(filePath) //nolint:gosec // G304 ignored cause we clean the path
	if err != nil {
		return nil, fmt.Errorf("failed to read screen state file: %w", err)
	}
	var state ScreenState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal screen state: %w", err)
	}

	return &state, nil
}

// SaveScreenState saves the state for a screen.
func (s *ScreenStorage) SaveScreenState(screenID string, state *ScreenState) error {
	f := filepath.Clean(filepath.Join(s.folder, screenID))
	filePath := filepath.Join(f, "state.json")
	xio.MakeDirAll(f)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal screen state: %w", err)
	}
	if err := xos.WriteFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write screen state to file: %w", err)
	}

	return nil
}
