package scenes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

const (
	// BattleMapVersion is the battle map format version this server reads.
	BattleMapVersion = 0
	// BattleMapExtension marks an asset reference as a battle map file instead of a plain image.
	BattleMapExtension = ".map.json"
	// DefaultCellSize is the size of a grid cell in pixels when a map shows a plain image.
	DefaultCellSize = 70
)

var (
	// ErrInvalidBattleMap is returned for battle map files that do not follow the format.
	ErrInvalidBattleMap = errors.New("invalid battle map")

	hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// MapLayerKind is the type of a battle map layer.
type MapLayerKind string

// Map layer kind constants.
const (
	MapLayerImage    MapLayerKind = "image"    // MapLayerImage draws a single image.
	MapLayerEntities MapLayerKind = "entities" // MapLayerEntities draws entity tokens and instance groups.
)

// BattleMap is the Mechanus battle map format: grid settings and an ordered list of typed layers.
// It is configuration, read from a file in the scene's folder and never modified at runtime: anything
// that moves during play belongs in [State].
type BattleMap struct {
	Version         int        `json:"version"`                    // Format version, see [BattleMapVersion]
	Name            string     `json:"name,omitempty"`             // Display name
	BackgroundColor string     `json:"background_color,omitempty"` // Hex color code behind all layers
	Grid            MapGrid    `json:"grid"`
	Layers          []MapLayer `json:"layers"` // Rendered in order, the first layer at the bottom
}

// MapGrid describes the grid of a battle map. One map pixel is one pixel of an image at its natural size.
type MapGrid struct {
	CellSize  float64 `json:"cell_size"`            // Size of one grid cell in map pixels
	Columns   int     `json:"columns"`              // Width of the map in grid cells
	Rows      int     `json:"rows"`                 // Height of the map in grid cells
	XOffset   int     `json:"x_offset,omitempty"`   // Moves the grid, and all that is placed on it, to the right, in map pixels
	YOffset   int     `json:"y_offset,omitempty"`   // Moves the grid, and all that is placed on it, down, in map pixels
	Hidden    bool    `json:"hidden,omitempty"`     // Do not draw the grid lines
	LineColor string  `json:"line_color,omitempty"` // Hex color code of the grid lines
}

// MapLayer is one layer of a battle map. Which fields apply depends on its kind.
type MapLayer struct {
	ID   string       `json:"id"` // Unique within the map
	Kind MapLayerKind `json:"kind"`

	// [MapLayerImage]
	Asset    string            `json:"asset,omitempty"`   // Image: scene relative or shared:/ reference
	Position vectors.Vector2D  `json:"position,omitzero"` // Top left corner in grid units, from the map's origin, not moved by the grid offset
	Size     *vectors.Vector2D `json:"size,omitempty"`    // Size in grid units, nil for the image's natural size
	Opacity  *float64          `json:"opacity,omitempty"` // Between 0.0 and 1.0, nil is fully opaque

	// [MapLayerEntities]
	Entities       []MapEntity     `json:"entities,omitempty"`        // Tokens of the scene's entities
	InstanceGroups []InstanceGroup `json:"instance_groups,omitempty"` // Scenery, one asset at many positions
}

// MapEntity places one of the scene's entities on the map. It is only drawn while that entity is present
// and visible; a removed, hidden, or unknown entity draws nothing. The position is where the entity
// starts: it is configuration, a position changed during play belongs in [State].
type MapEntity struct {
	Entity   string            `json:"entity"`             // ID of the [Entity] in the scene
	Asset    string            `json:"asset,omitempty"`    // Token image, empty for a placeholder token
	Position vectors.Vector2D  `json:"position"`           // Top left corner in grid units
	Size     *vectors.Vector2D `json:"size,omitempty"`     // Size in grid units, nil for one cell
	Rotation float64           `json:"rotation,omitempty"` // Rotation in degrees, around the center
}

// InstanceGroup draws one asset at many positions (trees, barrels, rubble), so a map can hold thousands
// of them. Instances are not entities: they have no identity in the scene.
type InstanceGroup struct {
	ID        string            `json:"id"`             // Unique within the layer
	Asset     string            `json:"asset"`          // Image of every instance
	Size      *vectors.Vector2D `json:"size,omitempty"` // Size of every instance in grid units, nil for one cell
	Instances []Instance        `json:"instances"`
}

// Instance is one position of an [InstanceGroup].
type Instance struct {
	X        float64 `json:"x"`                  // Left in grid units
	Y        float64 `json:"y"`                  // Top in grid units
	Rotation float64 `json:"rotation,omitempty"` // Rotation in degrees, around the center
}

// IsBattleMapFile reports whether an asset reference points to a battle map file.
func IsBattleMapFile(ref string) bool {
	return strings.HasSuffix(ref, BattleMapExtension)
}

// ParseBattleMap reads and validates a battle map file. Unknown fields are rejected, as these files
// are written by hand and a typo would otherwise silently do nothing.
func ParseBattleMap(r io.Reader) (*BattleMap, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var m BattleMap
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidBattleMap, err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return &m, nil
}

// ImageBattleMap creates the battle map for a plain image: a default grid over that image, sized by
// the image.
func ImageBattleMap(asset string) *BattleMap {
	return &BattleMap{
		Version: BattleMapVersion,
		Grid:    MapGrid{CellSize: DefaultCellSize},
		Layers:  []MapLayer{{ID: "image", Kind: MapLayerImage, Asset: asset}},
	}
}

// Validate checks the map follows the format.
func (m *BattleMap) Validate() error {
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if m.Version != BattleMapVersion {
		fail("unsupported version %d, expected %d", m.Version, BattleMapVersion)
	}
	if m.BackgroundColor != "" && !hexColor.MatchString(m.BackgroundColor) {
		fail("background_color %q is not a hex color", m.BackgroundColor)
	}
	errs = append(errs, m.Grid.validate()...)

	layerIDs := map[string]bool{}
	entities := map[string]bool{}
	for i := range m.Layers {
		l := &m.Layers[i]
		if l.ID == "" {
			fail("layer %d: empty id", i)
		}
		if layerIDs[l.ID] {
			fail("layer %q: duplicate id", l.ID)
		}
		layerIDs[l.ID] = true

		for _, err := range l.validate(entities) {
			fail("layer %q: %w", l.ID, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBattleMap, err)
	}

	return nil
}

func (g *MapGrid) validate() []error {
	var errs []error
	if g.CellSize <= 0 {
		errs = append(errs, fmt.Errorf("grid: cell_size must be positive, got %v", g.CellSize))
	}
	if g.Columns <= 0 || g.Rows <= 0 {
		errs = append(errs, fmt.Errorf("grid: columns and rows must be positive, got %d x %d", g.Columns, g.Rows))
	}
	if g.LineColor != "" && !hexColor.MatchString(g.LineColor) {
		errs = append(errs, fmt.Errorf("grid: line_color %q is not a hex color", g.LineColor))
	}

	return errs
}

func (l *MapLayer) validate(entities map[string]bool) []error {
	switch l.Kind {
	case MapLayerImage:
		return l.validateImage()
	case MapLayerEntities:
		return l.validateEntities(entities)
	default:
		return []error{fmt.Errorf("unknown kind %q", l.Kind)}
	}
}

func (l *MapLayer) validateImage() []error {
	var errs []error
	if l.Asset == "" {
		errs = append(errs, errors.New("image layer without asset"))
	}
	if l.Opacity != nil && (*l.Opacity < 0 || *l.Opacity > 1) {
		errs = append(errs, fmt.Errorf("opacity must be between 0 and 1, got %v", *l.Opacity))
	}
	if err := validateSize(l.Size); err != nil {
		errs = append(errs, err)
	}
	if len(l.Entities) > 0 || len(l.InstanceGroups) > 0 {
		errs = append(errs, errors.New("image layer with entities or instance groups"))
	}

	return errs
}

func (l *MapLayer) validateEntities(seen map[string]bool) []error {
	var errs []error
	if l.Asset != "" || l.Size != nil || l.Opacity != nil {
		errs = append(errs, errors.New("entities layer with image fields"))
	}

	for i := range l.Entities {
		e := &l.Entities[i]
		if e.Entity == "" {
			errs = append(errs, fmt.Errorf("entity %d: empty entity id", i))
		}
		if seen[e.Entity] {
			errs = append(errs, fmt.Errorf("entity %q: placed more than once on the map", e.Entity))
		}
		seen[e.Entity] = true

		if err := validateSize(e.Size); err != nil {
			errs = append(errs, fmt.Errorf("entity %q: %w", e.Entity, err))
		}
	}

	groups := map[string]bool{}
	for i := range l.InstanceGroups {
		g := &l.InstanceGroups[i]
		if g.ID == "" {
			errs = append(errs, fmt.Errorf("instance group %d: empty id", i))
		}
		if groups[g.ID] {
			errs = append(errs, fmt.Errorf("instance group %q: duplicate id", g.ID))
		}
		groups[g.ID] = true

		if g.Asset == "" {
			errs = append(errs, fmt.Errorf("instance group %q: no asset", g.ID))
		}
		if err := validateSize(g.Size); err != nil {
			errs = append(errs, fmt.Errorf("instance group %q: %w", g.ID, err))
		}
	}

	return errs
}

func validateSize(size *vectors.Vector2D) error {
	if size != nil && (size.X <= 0 || size.Y <= 0) {
		return fmt.Errorf("size must be positive, got %v x %v", size.X, size.Y)
	}

	return nil
}
