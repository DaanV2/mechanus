package state

import (
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

// View is what a screen shows, with the overlays drawn on top of it.
type View struct {
	Kind      scenes.ViewKind   `json:"kind"`
	Asset     *Asset            `json:"asset,omitempty"`     // Battle map, image, video, ... depending on the view
	Component string            `json:"component,omitempty"` // Client component name, for [scenes.ViewCustom]
	Params    map[string]string `json:"params,omitempty"`    // View specific options
	Overlays  []scenes.Overlay  `json:"overlays,omitempty"`  // Widgets drawn on top of the view, in order
}

// Camera is the part of the view that is visible on the screen.
type Camera struct {
	Center vectors.Vector2D `json:"center"` // Center of the screen, in grid units
	Zoom   float64          `json:"zoom"`   // Scale factor, 1.0 is the default size
}
