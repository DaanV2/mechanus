package scenestore

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
)

// StateStore persists the runtime state of scenes, one record per scene.
type StateStore struct {
	states storage.Storage[*scenes.State]
}

// NewStateStore creates a state store on top of the given storage.
func NewStateStore(stateStorage storage.Storage[*scenes.State]) *StateStore {
	return &StateStore{states: stateStorage}
}

// NewDBStateStore creates a state store backed by the database.
func NewDBStateStore(db *persistence.DB) (*StateStore, error) {
	stateStorage, err := storage.NamedDBStorage[*scenes.State](db, "scene").StateStorage()
	if err != nil {
		return nil, fmt.Errorf("scene state storage: %w", err)
	}

	return NewStateStore(stateStorage), nil
}

// LoadSceneState implements application.SceneStateStore.
func (s *StateStore) LoadSceneState(ctx context.Context, sceneID string) (*scenes.State, error) {
	state, err := s.states.Get(ctx, sceneID)
	if isNotFound(err) {
		return nil, fmt.Errorf("%q: %w", sceneID, scenes.ErrStateNotFound)
	}

	return state, err
}

// SaveSceneState implements application.SceneStateStore.
func (s *StateStore) SaveSceneState(ctx context.Context, state *scenes.State) error {
	return s.states.Set(ctx, state)
}
