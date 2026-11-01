package state

import (
	"maps"
	"slices"

	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var viewKinds = map[scenes.ViewKind]screensv1.ViewKind{
	scenes.ViewBattleMap: screensv1.ViewKind_VIEW_KIND_BATTLEMAP,
	scenes.ViewScenery:   screensv1.ViewKind_VIEW_KIND_SCENERY,
	scenes.ViewInfoBoard: screensv1.ViewKind_VIEW_KIND_INFOBOARD,
	scenes.ViewCustom:    screensv1.ViewKind_VIEW_KIND_CUSTOM,
	scenes.ViewSplash:    screensv1.ViewKind_VIEW_KIND_SPLASH,
	scenes.ViewBlank:     screensv1.ViewKind_VIEW_KIND_BLANK,
}

var overlayKinds = map[scenes.OverlayKind]screensv1.OverlayKind{
	scenes.OverlayInitiative: screensv1.OverlayKind_OVERLAY_KIND_INITIATIVE,
	scenes.OverlayEntities:   screensv1.OverlayKind_OVERLAY_KIND_ENTITIES,
	scenes.OverlayCustom:     screensv1.OverlayKind_OVERLAY_KIND_CUSTOM,
}

var entityKinds = map[scenes.EntityKind]screensv1.EntityKind{
	scenes.EntityPlayer: screensv1.EntityKind_ENTITY_KIND_PLAYER,
	scenes.EntityNPC:    screensv1.EntityKind_ENTITY_KIND_NPC,
	scenes.EntityEnemy:  screensv1.EntityKind_ENTITY_KIND_ENEMY,
	scenes.EntityObject: screensv1.EntityKind_ENTITY_KIND_OBJECT,
}

// ToProto converts the screen state into the message sent to screen clients.
func (s *ScreenState) ToProto() *screensv1.ScreenState {
	if s == nil {
		return nil
	}

	layers := make([]*screensv1.Layer, 0, len(s.Layers))
	for _, l := range s.Layers {
		if l != nil {
			layers = append(layers, l.ToProto())
		}
	}

	return &screensv1.ScreenState{
		Environment:  s.Environment.ToProto(),
		Grid:         s.Grid.ToProto(),
		Layer:        layers,
		View:         s.View.ToProto(),
		Camera:       s.Camera.ToProto(),
		SplashScreen: s.SplashScreen.ToProto(),
	}
}

// ToProto converts the environment into its proto message.
func (e *Environment) ToProto() *screensv1.Environment {
	if e == nil {
		return nil
	}

	return &screensv1.Environment{
		CurrentTime:     timestamppb.New(e.CurrentTime),
		BackgroundColor: e.BackgroundColor,
		LightIntensity:  e.LightIntensity,
	}
}

// ToProto converts the grid settings into its proto message.
func (g *GridSettings) ToProto() *screensv1.GridSettings {
	if g == nil {
		return nil
	}

	return &screensv1.GridSettings{
		XOffset:   int32(g.XOffset), //nolint:gosec // G115 pixel offsets fit in an int32
		YOffset:   int32(g.YOffset), //nolint:gosec // G115 pixel offsets fit in an int32
		XRatio:    int32(g.XRatio),  //nolint:gosec // G115 ratios fit in an int32
		YRatio:    int32(g.YRatio),  //nolint:gosec // G115 ratios fit in an int32
		Show:      g.Show,
		CellSize:  g.CellSize,
		LineColor: g.LineColor,
	}
}

// ToProto converts the layer into its proto message, entities are sorted by ID.
func (l *Layer) ToProto() *screensv1.Layer {
	if l == nil {
		return nil
	}

	entities := make([]*screensv1.Entity, 0, len(l.Entities))
	for _, id := range slices.Sorted(maps.Keys(l.Entities)) {
		if e := l.Entities[id]; e != nil {
			entities = append(entities, e.ToProto())
		}
	}

	return &screensv1.Layer{
		Id:     l.ID,
		Entity: entities,
	}
}

// ToProto converts the entity into its proto message.
func (e *Entity) ToProto() *screensv1.Entity {
	if e == nil {
		return nil
	}

	return &screensv1.Entity{
		Id:       e.ID,
		Name:     e.Name,
		Kind:     entityKinds[e.Kind],
		Asset:    e.Asset.ToProto(),
		Position: vectorToProto(e.Position),
		Size:     vectorToProto(e.Size),
		Rotation: e.Rotation,
		Hidden:   e.Hidden,
	}
}

// ToProto converts the asset into its proto message.
func (a *Asset) ToProto() *screensv1.Asset {
	if a == nil {
		return nil
	}

	return &screensv1.Asset{
		Id:        a.ID,
		Name:      a.Name,
		Url:       a.Url,
		MediaType: a.MediaType,
	}
}

// ToProto converts the view into its proto message.
func (v *View) ToProto() *screensv1.View {
	if v == nil {
		return nil
	}

	overlays := make([]*screensv1.Overlay, 0, len(v.Overlays))
	for _, o := range v.Overlays {
		overlays = append(overlays, &screensv1.Overlay{
			Kind:      overlayKinds[o.Kind],
			Component: optionalString(o.Component),
			Params:    o.Params,
		})
	}

	return &screensv1.View{
		Kind:      viewKinds[v.Kind],
		Asset:     v.Asset.ToProto(),
		Component: optionalString(v.Component),
		Params:    v.Params,
		Overlay:   overlays,
	}
}

// ToProto converts the camera into its proto message.
func (c *Camera) ToProto() *screensv1.Camera {
	if c == nil {
		return nil
	}

	return &screensv1.Camera{
		Center: vectorToProto(c.Center),
		Zoom:   c.Zoom,
	}
}

// ToProto converts the splash screen into its proto message.
func (s *SplashScreen) ToProto() *screensv1.SplashScreen {
	if s == nil {
		return nil
	}

	return &screensv1.SplashScreen{
		Show:          s.Show,
		Title:         new(s.Title),
		Subtitle:      new(s.Subtitle),
		BackgroundHex: new(s.BackgroundColor),
	}
}

func vectorToProto(v vectors.Vector2D) *screensv1.Vector2 {
	return &screensv1.Vector2{X: v.X, Y: v.Y}
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
