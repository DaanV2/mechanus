package scenes

import (
	"errors"
	"fmt"
)

var (
	// ErrStateNotFound is returned by state stores when a scene has never been saved.
	ErrStateNotFound = errors.New("scene state not found")
	// ErrEntityNotFound is returned when an entity is not part of the scene.
	ErrEntityNotFound = errors.New("entity not found")
	// ErrEntityExists is returned when spawning an entity whose ID is already in use.
	ErrEntityExists = errors.New("entity already exists")
)

// EntityState is the runtime status of an entity.
type EntityState struct {
	Hidden  bool `json:"hidden"`  // Hidden from players
	Removed bool `json:"removed"` // Defeated, fled, or otherwise no longer present
}

// State is everything about a scene that changes during play. It is created from a [Config] the
// first time the scene is loaded, saved to the database when the scene is exited, and restored
// when the scene is loaded again. It only stores what differs from or adds to the configuration.
type State struct {
	SceneID   string                 `json:"scene_id"`
	VariantID string                 `json:"variant_id,omitempty"` // Active variant, empty for base instructions
	Entities  map[string]EntityState `json:"entities"`             // Runtime status per entity ID (configured and spawned)
	Spawned   []Entity               `json:"spawned,omitempty"`    // Entities added during play that are not in the config
	Overrides Instructions           `json:"overrides"`            // Live GM adjustments on top of the configured instructions
}

// NewState creates the initial state for a scene from its configuration.
func NewState(cfg *Config) (*State, error) {
	variantID, err := cfg.ResolveVariantID("")
	if err != nil {
		return nil, err
	}

	s := &State{
		SceneID:   cfg.ID,
		VariantID: variantID,
		Entities:  make(map[string]EntityState, len(cfg.Entities)),
	}
	s.sync(cfg)

	return s, nil
}

// GetID returns the scene ID, states are stored per scene.
func (s *State) GetID() string {
	return s.SceneID
}

// SetHidden hides or reveals an entity.
func (s *State) SetHidden(entityID string, hidden bool) error {
	es, ok := s.Entities[entityID]
	if !ok {
		return fmt.Errorf("%q: %w", entityID, ErrEntityNotFound)
	}
	es.Hidden = hidden
	s.Entities[entityID] = es

	return nil
}

// Remove marks an entity as no longer present.
func (s *State) Remove(entityID string) error {
	es, ok := s.Entities[entityID]
	if !ok {
		return fmt.Errorf("%q: %w", entityID, ErrEntityNotFound)
	}
	es.Removed = true
	s.Entities[entityID] = es

	return nil
}

// Spawn adds an entity during play (summons, reinforcements, a surprise NPC).
func (s *State) Spawn(entity Entity) error {
	if _, ok := s.Entities[entity.ID]; ok {
		return fmt.Errorf("%q: %w", entity.ID, ErrEntityExists)
	}
	s.Spawned = append(s.Spawned, entity)
	s.Entities[entity.ID] = EntityState{Hidden: entity.Hidden}

	return nil
}

// sync adds entities that were added to the configuration after this state was saved.
func (s *State) sync(cfg *Config) {
	if s.Entities == nil {
		s.Entities = map[string]EntityState{}
	}
	for _, e := range cfg.Entities {
		if _, ok := s.Entities[e.ID]; !ok {
			s.Entities[e.ID] = EntityState{Hidden: e.Hidden}
		}
	}
}
