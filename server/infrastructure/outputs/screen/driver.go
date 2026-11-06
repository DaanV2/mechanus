// Package screen implements the output driver for the built-in browser screen client: it turns a
// scene's screen instruction into screen state and pushes it to the device's screen over the
// WebSocket.
package screen

import (
	"context"
	"fmt"
	"io/fs"
	"maps"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
)

var (
	_ application.OutputDriver    = &Driver{}
	_ application.BattleMapDriver = &Driver{}
)

// AssetResolver turns asset references from a scene config into URLs screens can load, and opens the
// ones the server reads itself.
type AssetResolver interface {
	Resolve(sceneID, ref string) (*assets.Resolved, error)
	Open(sceneID, ref string) (fs.File, error)
}

// Driver delivers screen instructions to screen devices. Each screen device has its own screen,
// keyed by device ID.
type Driver struct {
	screens *screens.ScreenManager
	assets  AssetResolver
}

// NewDriver creates a screen driver pushing state to the given screens.
func NewDriver(screenManager *screens.ScreenManager, resolver AssetResolver) *Driver {
	return &Driver{screens: screenManager, assets: resolver}
}

// Integration implements application.OutputDriver.
func (d *Driver) Integration() string {
	return devices.IntegrationScreen
}

// Apply implements application.OutputDriver. Instructions of other kinds are ignored.
func (d *Driver) Apply(ctx context.Context, scene scenes.Ref, _ string, device *devices.Device, instruction scenes.Instruction) error {
	ins, ok := instruction.(scenes.ScreenInstruction)
	if !ok {
		return nil
	}

	st, err := d.State(scene, &ins)
	if err != nil {
		return err
	}
	d.screens.GetOrCreate(device.ID).SetState(ctx, st)

	return nil
}

// State builds the screen state for a screen instruction of the given scene.
func (d *Driver) State(scene scenes.Ref, ins *scenes.ScreenInstruction) (*state.ScreenState, error) {
	st := state.NewScreenState()
	st.View = &state.View{
		Kind:      ins.View,
		Component: ins.Component,
		Params:    ins.Params,
		Overlays:  ins.Overlays,
	}
	if ins.View == "" {
		st.View.Kind = scenes.ViewBlank
	}

	if ins.View == scenes.ViewBattleMap && ins.Camera != nil {
		st.Camera = &state.Camera{Center: ins.Camera.Center, Zoom: ins.Camera.Zoom}
	}
	if ins.View == scenes.ViewBattleMap {
		if err := d.applyBattleMap(st, scene.SceneID, ins); err != nil {
			return nil, err
		}

		return st, nil
	}

	if ins.Asset != "" {
		asset, err := d.asset(scene.SceneID, ins.Asset)
		if err != nil {
			return nil, fmt.Errorf("screen view asset: %w", err)
		}
		st.View.Asset = asset
	}

	return st, nil
}

// applyBattleMap adds the battle map the instruction points to: a battle map file, or a plain image
// shown under a default grid. Without an asset the battle map is empty.
func (d *Driver) applyBattleMap(st *state.ScreenState, sceneID string, ins *scenes.ScreenInstruction) error {
	if ins.Asset == "" {
		return nil
	}

	m := scenes.ImageBattleMap(ins.Asset)
	if scenes.IsBattleMapFile(ins.Asset) {
		var err error
		if m, err = d.loadBattleMap(sceneID, ins.Asset); err != nil {
			return fmt.Errorf("battle map: %w", err)
		}
	} else {
		asset, err := d.asset(sceneID, ins.Asset)
		if err != nil {
			return fmt.Errorf("screen view asset: %w", err)
		}
		st.View.Asset = asset
	}

	if m.BackgroundColor != "" {
		st.Environment.BackgroundColor = m.BackgroundColor
	}
	st.Grid.CellSize = m.Grid.CellSize
	st.Grid.Columns = m.Grid.Columns
	st.Grid.Rows = m.Grid.Rows
	st.Grid.XOffset = m.Grid.XOffset
	st.Grid.YOffset = m.Grid.YOffset
	st.Grid.Show = !m.Grid.Hidden
	if m.Grid.LineColor != "" {
		st.Grid.LineColor = m.Grid.LineColor
	}

	present := make(map[string]*scenes.Entity, len(ins.Entities))
	for i := range ins.Entities {
		present[ins.Entities[i].ID] = &ins.Entities[i]
	}

	unplaced := unplacedTokens(m, ins)
	for i := range m.Layers {
		layer, err := d.layer(sceneID, &m.Layers[i], present, ins.Positions)
		if err != nil {
			return fmt.Errorf("battle map layer %q: %w", m.Layers[i].ID, err)
		}
		// Entities the map doesn't place, but that were given a position during play (spawned ones), go
		// on the first entity layer
		if layer.Kind == scenes.MapLayerEntities && unplaced != nil {
			maps.Copy(layer.Entities, unplaced)
			unplaced = nil
		}
		st.Layers = append(st.Layers, layer)
	}
	if unplaced != nil {
		st.Layers = append(st.Layers, &state.Layer{ID: PlayLayerID, Kind: scenes.MapLayerEntities, Entities: unplaced})
	}

	return nil
}

// PlayLayerID is the ID of the entity layer added to a battle map without one, to draw the entities
// that were given a position during play.
const PlayLayerID = "play"

// unplacedTokens returns placeholder tokens of the visible entities the map doesn't place but that have
// a position in the state, nil when there are none.
func unplacedTokens(m *scenes.BattleMap, ins *scenes.ScreenInstruction) map[string]*state.Entity {
	placed := map[string]bool{}
	for i := range m.Layers {
		for _, e := range m.Layers[i].Entities {
			placed[e.Entity] = true
		}
	}

	var result map[string]*state.Entity
	for _, e := range ins.Entities {
		pos, ok := ins.Positions[e.ID]
		if !ok || placed[e.ID] {
			continue
		}
		if result == nil {
			result = map[string]*state.Entity{}
		}
		result[e.ID] = &state.Entity{
			ID:       e.ID,
			Name:     e.Name,
			Kind:     e.Kind,
			Position: pos,
			Size:     vectors.Vector2D{X: 1, Y: 1},
		}
	}

	return result
}

// MoveEntity implements application.BattleMapDriver: the device's screen gets only the moved token.
func (d *Driver) MoveEntity(ctx context.Context, _ scenes.Ref, device *devices.Device, entityID string, position vectors.Vector2D) error {
	if screen, ok := d.screens.Get(device.ID); ok {
		screen.MoveEntity(ctx, entityID, position)
	}

	return nil
}

// SetCamera implements application.BattleMapDriver: the device's screen gets only the camera.
func (d *Driver) SetCamera(ctx context.Context, _ scenes.Ref, device *devices.Device, camera scenes.Camera) error {
	d.screens.GetOrCreate(device.ID).SetCamera(ctx, state.Camera{Center: camera.Center, Zoom: camera.Zoom})

	return nil
}

func (d *Driver) loadBattleMap(sceneID, ref string) (*scenes.BattleMap, error) {
	f, err := d.assets.Open(sceneID, ref)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	m, err := scenes.ParseBattleMap(f)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", ref, err)
	}

	return m, nil
}

func (d *Driver) layer(sceneID string, l *scenes.MapLayer, present map[string]*scenes.Entity, positions map[string]vectors.Vector2D) (*state.Layer, error) {
	result := &state.Layer{ID: l.ID, Kind: l.Kind}

	if l.Kind == scenes.MapLayerImage {
		asset, err := d.asset(sceneID, l.Asset)
		if err != nil {
			return nil, err
		}
		result.Asset = asset
		result.Position = l.Position
		result.Size = l.Size
		result.Opacity = 1
		if l.Opacity != nil {
			result.Opacity = *l.Opacity
		}

		return result, nil
	}

	result.Entities = make(map[string]*state.Entity, len(l.Entities))
	for i := range l.Entities {
		placed := &l.Entities[i]
		entity, ok := present[placed.Entity]
		if !ok {
			continue
		}

		token := &state.Entity{
			ID:       entity.ID,
			Name:     entity.Name,
			Kind:     entity.Kind,
			Position: placed.Position,
			Size:     cellsOrOne(placed.Size),
			Rotation: placed.Rotation,
		}
		if pos, ok := positions[entity.ID]; ok {
			token.Position = pos
		}
		if placed.Asset != "" {
			asset, err := d.asset(sceneID, placed.Asset)
			if err != nil {
				return nil, fmt.Errorf("entity %q: %w", placed.Entity, err)
			}
			token.Asset = asset
		}
		result.Entities[entity.ID] = token
	}

	for i := range l.InstanceGroups {
		g := &l.InstanceGroups[i]
		asset, err := d.asset(sceneID, g.Asset)
		if err != nil {
			return nil, fmt.Errorf("instance group %q: %w", g.ID, err)
		}
		result.InstanceGroups = append(result.InstanceGroups, &state.InstanceGroup{
			ID:        g.ID,
			Asset:     asset,
			Size:      cellsOrOne(g.Size),
			Instances: g.Instances,
		})
	}

	return result, nil
}

func (d *Driver) asset(sceneID, ref string) (*state.Asset, error) {
	asset, err := d.assets.Resolve(sceneID, ref)
	if err != nil {
		return nil, err
	}

	return &state.Asset{ID: ref, Name: ref, Url: asset.URL, MediaType: asset.MediaType}, nil
}

func cellsOrOne(size *vectors.Vector2D) vectors.Vector2D {
	if size == nil {
		return vectors.Vector2D{X: 1, Y: 1}
	}

	return *size
}
