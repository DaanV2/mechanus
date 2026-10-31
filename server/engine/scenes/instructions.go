package scenes

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"time"

	"github.com/DaanV2/mechanus/server/engine/devices"
)

// Instruction is something a device of a specific kind should do.
type Instruction interface {
	Kind() devices.Kind
}

// ViewKind is the type of view a screen shows.
type ViewKind string

// View kind constants.
const (
	ViewBattleMap ViewKind = "battlemap" // ViewBattleMap is a tactical grid map with entities, in Mechanus' own map format.
	ViewScenery   ViewKind = "scenery"   // ViewScenery is an image / video backdrop.
	ViewInfoBoard ViewKind = "infoboard" // ViewInfoBoard shows text / handouts / initiative.
	ViewCustom    ViewKind = "custom"    // ViewCustom is a client side component chosen by name.
	ViewSplash    ViewKind = "splash"    // ViewSplash is a title card / loading screen.
	ViewBlank     ViewKind = "blank"     // ViewBlank turns the screen black.
)

// OverlayKind is the type of widget drawn on top of a view.
type OverlayKind string

// Overlay kind constants.
const (
	OverlayInitiative OverlayKind = "initiative" // OverlayInitiative is an initiative / turn order tracker.
	OverlayEntities   OverlayKind = "entities"   // OverlayEntities shows the scene's visible entities.
	OverlayCustom     OverlayKind = "custom"     // OverlayCustom is a client side component chosen by name.
)

// Overlay is a widget drawn on top of a screen's view.
type Overlay struct {
	Kind      OverlayKind       `json:"kind"`
	Component string            `json:"component,omitempty"` // Client component name, for [OverlayCustom]
	Params    map[string]string `json:"params,omitempty"`    // Overlay specific options (position, size, ...)
}

// ScreenInstruction tells a screen which view to show.
type ScreenInstruction struct {
	View      ViewKind          `json:"view"`
	Asset     string            `json:"asset,omitempty"`     // Battle map, image, video, ... depending on the view
	Component string            `json:"component,omitempty"` // Client component name, for [ViewCustom]
	Params    map[string]string `json:"params,omitempty"`    // View specific options
	Overlays  []Overlay         `json:"overlays,omitempty"`  // Widgets drawn on top of the view, in order
}

// Kind implements [Instruction].
func (ScreenInstruction) Kind() devices.Kind { return devices.KindScreen }

// LightInstruction tells a light (group) what to do.
type LightInstruction struct {
	On         bool          `json:"on"`
	Color      string        `json:"color,omitempty"`  // Hex color code, e.g. "#ff8800"
	Brightness float64       `json:"brightness"`       // Value between 0.0 (off) and 1.0 (full)
	Effect     string        `json:"effect,omitempty"` // Integration specific effect, e.g. "candle", "lightning"
	Fade       time.Duration `json:"fade,omitempty"`   // Time to transition into this state
}

// Kind implements [Instruction].
func (LightInstruction) Kind() devices.Kind { return devices.KindLight }

// MusicAction is what a music player should do.
type MusicAction string

// Music action constants.
const (
	MusicPlay  MusicAction = "play"
	MusicPause MusicAction = "pause"
	MusicStop  MusicAction = "stop"
)

// MusicInstruction tells a music player what to play.
type MusicInstruction struct {
	Action    MusicAction   `json:"action"`
	Source    string        `json:"source,omitempty"` // Integration specific URI, e.g. "spotify:playlist:..."
	Volume    float64       `json:"volume"`           // Value between 0.0 and 1.0
	Shuffle   bool          `json:"shuffle,omitempty"`
	Loop      bool          `json:"loop,omitempty"`
	CrossFade time.Duration `json:"cross_fade,omitempty"`
}

// Kind implements [Instruction].
func (MusicInstruction) Kind() devices.Kind { return devices.KindMusic }

// CustomInstruction is a raw payload for a custom device, delivered over an event bus.
type CustomInstruction struct {
	Topic   string          `json:"topic"`   // Event bus topic / channel, e.g. an MQTT topic
	Payload json.RawMessage `json:"payload"` // Opaque payload, passed through untouched
}

// Kind implements [Instruction].
func (CustomInstruction) Kind() devices.Kind { return devices.KindCustom }

// Instructions holds what every output should do, keyed by target name per device kind.
type Instructions struct {
	Screens map[string]ScreenInstruction `json:"screens,omitempty"`
	Lights  map[string]LightInstruction  `json:"lights,omitempty"`
	Music   map[string]MusicInstruction  `json:"music,omitempty"`
	Custom  map[string]CustomInstruction `json:"custom,omitempty"`
}

// TargetRef identifies a target for a specific device kind.
type TargetRef struct {
	Kind   devices.Kind `json:"kind"`
	Target string       `json:"target"`
}

// TargetedInstruction is an instruction together with the target it is addressed to.
type TargetedInstruction struct {
	Target      string
	Instruction Instruction
}

// Merge returns a new set of instructions where overrides replace the instruction of the same
// target and kind. Neither input is modified.
func (i Instructions) Merge(overrides Instructions) Instructions {
	return Instructions{
		Screens: mergeMap(i.Screens, overrides.Screens),
		Lights:  mergeMap(i.Lights, overrides.Lights),
		Music:   mergeMap(i.Music, overrides.Music),
		Custom:  mergeMap(i.Custom, overrides.Custom),
	}
}

// Changes returns the instructions in next that are new or differ from i.
func (i Instructions) Changes(next Instructions) Instructions {
	return Instructions{
		Screens: changedMap(i.Screens, next.Screens),
		Lights:  changedMap(i.Lights, next.Lights),
		Music:   changedMap(i.Music, next.Music),
		Custom:  changedMap(i.Custom, next.Custom),
	}
}

// Released returns the targets that have an instruction in i but none in next.
func (i Instructions) Released(next Instructions) []TargetRef {
	var result []TargetRef
	result = appendReleased(result, devices.KindScreen, i.Screens, next.Screens)
	result = appendReleased(result, devices.KindLight, i.Lights, next.Lights)
	result = appendReleased(result, devices.KindMusic, i.Music, next.Music)
	result = appendReleased(result, devices.KindCustom, i.Custom, next.Custom)

	return result
}

// IsEmpty returns true when there are no instructions at all.
func (i Instructions) IsEmpty() bool {
	return len(i.Screens) == 0 && len(i.Lights) == 0 && len(i.Music) == 0 && len(i.Custom) == 0
}

// All returns every instruction with its target, sorted by kind then target for stable dispatch.
func (i Instructions) All() []TargetedInstruction {
	var result []TargetedInstruction
	result = appendTargeted(result, i.Screens)
	result = appendTargeted(result, i.Lights)
	result = appendTargeted(result, i.Music)
	result = appendTargeted(result, i.Custom)

	return result
}

func mergeMap[T any](base, overrides map[string]T) map[string]T {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}
	result := make(map[string]T, len(base)+len(overrides))
	maps.Copy(result, base)
	maps.Copy(result, overrides)

	return result
}

func changedMap[T any](prev, next map[string]T) map[string]T {
	var result map[string]T
	for target, ins := range next {
		if old, ok := prev[target]; ok && reflect.DeepEqual(old, ins) {
			continue
		}
		if result == nil {
			result = map[string]T{}
		}
		result[target] = ins
	}

	return result
}

func appendReleased[T any](result []TargetRef, kind devices.Kind, prev, next map[string]T) []TargetRef {
	for _, target := range slices.Sorted(maps.Keys(prev)) {
		if _, ok := next[target]; !ok {
			result = append(result, TargetRef{Kind: kind, Target: target})
		}
	}

	return result
}

func appendTargeted[T Instruction](result []TargetedInstruction, m map[string]T) []TargetedInstruction {
	for _, target := range slices.Sorted(maps.Keys(m)) {
		result = append(result, TargetedInstruction{Target: target, Instruction: m[target]})
	}

	return result
}
