package state

import (
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

// Layer represents a rendering layer. Which fields apply depends on its kind.
type Layer struct {
	ID   string              `json:"id"`   // Unique identifier for the layer within the screen
	Kind scenes.MapLayerKind `json:"kind"` // Image or entities

	// [scenes.MapLayerImage]
	Asset    *Asset            `json:"asset,omitempty"`   // The image
	Position vectors.Vector2D  `json:"position,omitzero"` // Top left corner in grid units, from the map's origin
	Size     *vectors.Vector2D `json:"size,omitempty"`    // Size in grid units, nil for the image's natural size
	Opacity  float64           `json:"opacity,omitempty"` // Value between 0.0 (invisible) and 1.0 (opaque)

	// [scenes.MapLayerEntities]
	Entities       map[string]*Entity `json:"entities,omitempty"`        // Map of entity IDs to Entity objects
	InstanceGroups []*InstanceGroup   `json:"instance_groups,omitempty"` // One asset drawn at many positions
}

// InstanceGroup is one asset drawn at many positions.
type InstanceGroup struct {
	ID        string            `json:"id"`        // Unique identifier for the group within the layer
	Asset     *Asset            `json:"asset"`     // Image of every instance
	Size      vectors.Vector2D  `json:"size"`      // Size of every instance in grid units
	Instances []scenes.Instance `json:"instances"` // The positions, in grid units
}
