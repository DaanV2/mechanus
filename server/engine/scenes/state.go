package scenes

import (
	"errors"
	"fmt"
	"math"

	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

const (
	// MaxCoordinate is the largest distance from the map's origin, in grid units, that a position or a
	// camera center may have. It keeps values sent by screens to something a map can show.
	MaxCoordinate = 10000
	// MinZoom is the smallest camera zoom factor.
	MinZoom = 0.1
	// MaxZoom is the largest camera zoom factor.
	MaxZoom = 10
)

var (
	// ErrStateNotFound is returned by state stores when a scene has never been saved.
	ErrStateNotFound = errors.New("scene state not found")
	// ErrEntityNotFound is returned when an entity is not part of the scene.
	ErrEntityNotFound = errors.New("entity not found")
	// ErrEntityExists is returned when spawning an entity whose ID is already in use.
	ErrEntityExists = errors.New("entity already exists")
	// ErrInvalidEntity is returned when spawning an entity without an ID or with an unknown kind.
	ErrInvalidEntity = errors.New("invalid entity")
	// ErrInvalidPosition is returned for a position that is not finite or further than [MaxCoordinate] from the origin.
	ErrInvalidPosition = errors.New("invalid position")
	// ErrInvalidCamera is returned for a camera with an invalid center, or a zoom outside [MinZoom, MaxZoom].
	ErrInvalidCamera = errors.New("invalid camera")
)

// EntityState is the runtime status of an entity.
type EntityState struct {
	Hidden   bool              `json:"hidden"`             // Hidden from players
	Removed  bool              `json:"removed"`            // Defeated, fled, or otherwise no longer present
	Position *vectors.Vector2D `json:"position,omitempty"` // Where it was moved to during play, in grid units. Nil keeps the position from the battle map
}

// Camera is the part of a battle map a screen shows.
type Camera struct {
	Center vectors.Vector2D `json:"center"` // Center of the screen, in grid units
	Zoom   float64          `json:"zoom"`   // Scale factor, 1.0 fits the whole map on the screen
}

// Validate checks the camera has a valid center and a zoom within [MinZoom, MaxZoom].
func (c Camera) Validate() error {
	if err := ValidatePosition(c.Center); err != nil {
		return fmt.Errorf("%w: center: %w", ErrInvalidCamera, err)
	}
	if math.IsNaN(c.Zoom) || c.Zoom < MinZoom || c.Zoom > MaxZoom {
		return fmt.Errorf("%w: zoom %v is not between %v and %v", ErrInvalidCamera, c.Zoom, MinZoom, MaxZoom)
	}

	return nil
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
	Cameras   map[string]Camera      `json:"cameras,omitempty"`    // Camera per screen target, a target without one shows the whole map
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

// Move places an entity at a position on the battle map, in grid units. A removed entity can't be moved.
func (s *State) Move(entityID string, position vectors.Vector2D) error {
	es, ok := s.Entities[entityID]
	if !ok || es.Removed {
		return fmt.Errorf("%q: %w", entityID, ErrEntityNotFound)
	}
	if err := ValidatePosition(position); err != nil {
		return fmt.Errorf("%q: %w", entityID, err)
	}
	es.Position = &position
	s.Entities[entityID] = es

	return nil
}

// SetCamera sets the camera of a screen target.
func (s *State) SetCamera(target string, camera Camera) error {
	if err := camera.Validate(); err != nil {
		return err
	}
	if s.Cameras == nil {
		s.Cameras = map[string]Camera{}
	}
	s.Cameras[target] = camera

	return nil
}

// Spawn adds an entity during play (summons, reinforcements, a surprise NPC).
func (s *State) Spawn(entity Entity) error {
	if err := entity.validate(); err != nil {
		return err
	}
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

// ValidatePosition checks a position is finite and no further than [MaxCoordinate] from the origin.
func ValidatePosition(p vectors.Vector2D) error {
	for _, v := range []float64{p.X, p.Y} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > MaxCoordinate {
			return fmt.Errorf("%w: (%v, %v)", ErrInvalidPosition, p.X, p.Y)
		}
	}

	return nil
}
