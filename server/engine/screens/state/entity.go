package state

import "github.com/DaanV2/mechanus/server/pkg/math/vectors"

// Entity represents an object in a screen that can be positioned and rendered.
type Entity struct {
	ID       string           `json:"id"`       // Unique identifier for the entity
	Name     string           `json:"name"`     // Name of the entity
	Position vectors.Vector2D `json:"position"` // Position of the entity in the grid
	Size     vectors.Vector2D `json:"size"`     // Size of the entity
	Rotation float64          `json:"rotation"` // Rotation of the entity in degrees
	Visible  bool             `json:"visible"`  // Whether the entity is visible
}
