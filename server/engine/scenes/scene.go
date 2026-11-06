package scenes

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

var (
	// ErrStateMismatch is returned when a state does not belong to the given configuration.
	ErrStateMismatch = errors.New("state does not belong to scene")
)

// Scene is a loaded scene: its read-only [Config] together with its runtime [State].
type Scene struct {
	Config *Config
	State  *State
}

// Load combines a configuration with its saved state. A nil state creates a fresh one from the
// configuration. Entities added to the configuration since the state was saved are picked up; a
// saved variant that no longer exists falls back to the default variant.
func Load(cfg *Config, state *State) (*Scene, error) {
	if state == nil {
		s, err := NewState(cfg)
		if err != nil {
			return nil, err
		}

		return &Scene{Config: cfg, State: s}, nil
	}
	if state.SceneID != cfg.ID {
		return nil, fmt.Errorf("state %q, scene %q: %w", state.SceneID, cfg.ID, ErrStateMismatch)
	}

	state.sync(cfg)
	if _, err := cfg.ResolveVariantID(state.VariantID); err != nil {
		id, err := cfg.ResolveVariantID("")
		if err != nil {
			return nil, err
		}
		state.VariantID = id
	}

	return &Scene{Config: cfg, State: state}, nil
}

// ID returns the scene ID.
func (s *Scene) ID() string {
	return s.Config.ID
}

// Ref returns the scene and active variant.
func (s *Scene) Ref() Ref {
	return Ref{SceneID: s.Config.ID, VariantID: s.State.VariantID}
}

// Instructions returns the effective instructions: the configured instructions for the active
// variant, with the live overrides from the state applied.
func (s *Scene) Instructions() (Instructions, error) {
	ins, err := s.Config.Resolve(s.State.VariantID)
	if err != nil {
		return Instructions{}, err
	}

	ins = ins.Merge(s.State.Overrides)
	for target, screen := range ins.Screens {
		if screen.View != ViewBattleMap {
			continue
		}
		screen.Entities = s.VisibleEntities()
		for _, e := range screen.Entities {
			if pos, ok := s.Position(e.ID); ok {
				if screen.Positions == nil {
					screen.Positions = map[string]vectors.Vector2D{}
				}
				screen.Positions[e.ID] = pos
			}
		}
		if camera, ok := s.State.Cameras[target]; ok {
			screen.Camera = &camera
		}
		ins.Screens[target] = screen
	}

	return ins, nil
}

// BattleMapTargets returns the screen targets that show a battle map, sorted.
func (s *Scene) BattleMapTargets() ([]string, error) {
	ins, err := s.Instructions()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, target := range slices.Sorted(maps.Keys(ins.Screens)) {
		if ins.Screens[target].View == ViewBattleMap {
			result = append(result, target)
		}
	}

	return result, nil
}

// Position returns where an entity was moved to during play. False when it was never moved, so it
// stands where the battle map places it.
func (s *Scene) Position(entityID string) (vectors.Vector2D, bool) {
	es, ok := s.State.Entities[entityID]
	if !ok || es.Position == nil {
		return vectors.Vector2D{}, false
	}

	return *es.Position, true
}

// SetVariant switches the active variant (empty for the default). Live overrides are cleared, as
// switching variants is a deliberate change of the whole scene's look.
func (s *Scene) SetVariant(variantID string) error {
	id, err := s.Config.ResolveVariantID(variantID)
	if err != nil {
		return err
	}
	s.State.VariantID = id
	s.State.Overrides = Instructions{}

	return nil
}

// Adjust applies live overrides on top of the current instructions, e.g. the GM turning down the
// music or changing a light colour mid-scene.
func (s *Scene) Adjust(overrides Instructions) {
	s.State.Overrides = s.State.Overrides.Merge(overrides)
}

// Entities returns the entities currently present: configured and spawned entities that have not
// been removed, with their runtime hidden status.
func (s *Scene) Entities() []Entity {
	var result []Entity
	add := func(e Entity) {
		es, ok := s.State.Entities[e.ID]
		if !ok || es.Removed {
			return
		}
		e.Hidden = es.Hidden
		result = append(result, e)
	}

	for _, e := range s.Config.Entities {
		add(e)
	}
	for _, e := range s.State.Spawned {
		add(e)
	}

	return result
}

// VisibleEntities returns the present entities that are not hidden from players.
func (s *Scene) VisibleEntities() []Entity {
	var result []Entity
	for _, e := range s.Entities() {
		if !e.Hidden {
			result = append(result, e)
		}
	}

	return result
}
