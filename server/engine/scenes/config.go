package scenes

import (
	"errors"
	"fmt"
)

var (
	// ErrSceneNotFound is returned by config stores when no scene with the given ID is configured.
	ErrSceneNotFound = errors.New("scene not found")
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
	Hidden      bool       `json:"hidden"`                 // In a [Config]: hidden when the scene is first loaded. Otherwise: currently hidden.
}

// Config is the authored definition of a scene, as prepared by the GM. It is loaded from file
// storage and never modified at runtime: everything that changes during play lives in [State].
type Config struct {
	ID             string       `json:"id"`
	CampaignID     string       `json:"campaign_id"`
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	Entities       []Entity     `json:"entities"`                  // Who / what is present when the scene starts
	Instructions   Instructions `json:"instructions"`              // Base instructions for every output
	Variants       []Variant    `json:"variants,omitempty"`        // Alternative states of this scene
	DefaultVariant string       `json:"default_variant,omitempty"` // Variant applied on first load, empty for base only
}

// Variant is an alternative state of a scene (time of day, weather, combat, ...). Its instructions
// override the scene's base instructions per target. Variants are part of the configuration.
type Variant struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Tags         map[string]string `json:"tags,omitempty"` // Free form classification, e.g. {"time_of_day": "night"}
	Instructions Instructions      `json:"instructions"`   // Overrides on top of the scene's base instructions
}

// GetID returns the scene ID.
func (c *Config) GetID() string {
	return c.ID
}

// Variant returns the variant with the given ID.
func (c *Config) Variant(id string) (*Variant, bool) {
	for i := range c.Variants {
		if c.Variants[i].ID == id {
			return &c.Variants[i], true
		}
	}

	return nil, false
}

// ResolveVariantID returns the variant that is applied when variantID is requested: the requested
// one, or the default when empty. Returns [ErrVariantNotFound] if it does not exist.
func (c *Config) ResolveVariantID(variantID string) (string, error) {
	if variantID == "" {
		variantID = c.DefaultVariant
	}
	if variantID == "" {
		return "", nil
	}
	if _, ok := c.Variant(variantID); !ok {
		return "", fmt.Errorf("scene %q variant %q: %w", c.ID, variantID, ErrVariantNotFound)
	}

	return variantID, nil
}

// Resolve returns the configured instructions for the given variant: the base instructions with
// the variant's overrides applied. An empty variantID uses [Config.DefaultVariant].
func (c *Config) Resolve(variantID string) (Instructions, error) {
	id, err := c.ResolveVariantID(variantID)
	if err != nil {
		return Instructions{}, err
	}
	if id == "" {
		return c.Instructions.Merge(Instructions{}), nil
	}

	v, _ := c.Variant(id)

	return c.Instructions.Merge(v.Instructions), nil
}
