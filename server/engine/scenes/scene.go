package scenes

import (
	"errors"
	"fmt"
)

var (
	// ErrVariantNotFound is returned when a scene does not have the requested variant.
	ErrVariantNotFound = errors.New("variant not found")
)

// EntityKind is the role an entity plays in a scene.
type EntityKind string

// Entity kind constants.
const (
	EntityPlayer EntityKind = "player" // EntityPlayer is a player character.
	EntityNPC    EntityKind = "npc"    // EntityNPC is a non hostile non-player character.
	EntityEnemy  EntityKind = "enemy"  // EntityEnemy is a hostile creature.
	EntityObject EntityKind = "object" // EntityObject is an interactable object (chest, door, trap, ...).
)

// Entity is something present in a scene. It is independent of how a screen renders it; views such
// as the battle map decide position and appearance.
type Entity struct {
	ID          string     `json:"id"`                     // Unique identifier within the scene
	Kind        EntityKind `json:"kind"`                   // Player, NPC, enemy, object
	Name        string     `json:"name"`                   // Display name
	CharacterID string     `json:"character_id,omitempty"` // Optional link to a persisted character
	Hidden      bool       `json:"hidden"`                 // Hidden from players until revealed by the GM
}

// Scene is a place or moment in the game, and the instructions that bring it to life in the room.
type Scene struct {
	ID             string       `json:"id"`
	CampaignID     string       `json:"campaign_id"`
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	Entities       []Entity     `json:"entities"`                  // Who / what is present
	Instructions   Instructions `json:"instructions"`              // Base instructions for every output
	Variants       []Variant    `json:"variants,omitempty"`        // Alternative states of this scene
	DefaultVariant string       `json:"default_variant,omitempty"` // Variant applied when none is requested, empty for base only
}

// Variant is an alternative state of a scene (time of day, weather, combat, ...). Its instructions
// override the scene's base instructions per target.
type Variant struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Tags         map[string]string `json:"tags,omitempty"` // Free form classification, e.g. {"time_of_day": "night"}
	Instructions Instructions      `json:"instructions"`   // Overrides on top of the scene's base instructions
}

// Variant returns the variant with the given ID.
func (s *Scene) Variant(id string) (*Variant, bool) {
	for i := range s.Variants {
		if s.Variants[i].ID == id {
			return &s.Variants[i], true
		}
	}

	return nil, false
}

// ResolveVariantID returns the variant that is applied when variantID is requested: the requested
// one, or the default when empty. Returns [ErrVariantNotFound] if it does not exist.
func (s *Scene) ResolveVariantID(variantID string) (string, error) {
	if variantID == "" {
		variantID = s.DefaultVariant
	}
	if variantID == "" {
		return "", nil
	}
	if _, ok := s.Variant(variantID); !ok {
		return "", fmt.Errorf("scene %q variant %q: %w", s.ID, variantID, ErrVariantNotFound)
	}

	return variantID, nil
}

// Resolve returns the effective instructions for the given variant: the base instructions with the
// variant's overrides applied. An empty variantID uses [Scene.DefaultVariant].
func (s *Scene) Resolve(variantID string) (Instructions, error) {
	id, err := s.ResolveVariantID(variantID)
	if err != nil {
		return Instructions{}, err
	}
	if id == "" {
		return s.Instructions.Merge(Instructions{}), nil
	}

	v, _ := s.Variant(id)

	return s.Instructions.Merge(v.Instructions), nil
}

// VisibleEntities returns the entities that are not hidden from players.
func (s *Scene) VisibleEntities() []Entity {
	var result []Entity
	for _, e := range s.Entities {
		if !e.Hidden {
			result = append(result, e)
		}
	}

	return result
}
