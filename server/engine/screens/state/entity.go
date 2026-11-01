package state

import (
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

// Entity represents an object in a screen that can be positioned and rendered.
type Entity struct {
	ID       string            `json:"id"`              // Unique identifier for the entity
	Name     string            `json:"name"`            // Name of the entity
	Kind     scenes.EntityKind `json:"kind"`            // Player, NPC, enemy, object
	Asset    *Asset            `json:"asset,omitempty"` // Token / image to render
	Position vectors.Vector2D  `json:"position"`        // Position of the entity in grid units
	Size     vectors.Vector2D  `json:"size"`            // Size of the entity in grid units
	Rotation float64           `json:"rotation"`        // Rotation of the entity in degrees
	Hidden   bool              `json:"hidden"`          // Hidden from players
}
