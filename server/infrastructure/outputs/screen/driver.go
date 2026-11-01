// Package screen implements the output driver for the built-in browser screen client: it turns a
// scene's screen instruction into screen state and pushes it to the device's screen over the
// WebSocket.
package screen

import (
	"context"
	"fmt"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
)

var _ application.OutputDriver = &Driver{}

// AssetResolver turns asset references from a scene config into URLs screens can load.
type AssetResolver interface {
	Resolve(sceneID, ref string) (*assets.Resolved, error)
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
func (d *Driver) Apply(ctx context.Context, scene scenes.Ref, device *devices.Device, instruction scenes.Instruction) error {
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

	if ins.Asset != "" {
		asset, err := d.assets.Resolve(scene.SceneID, ins.Asset)
		if err != nil {
			return nil, fmt.Errorf("screen view asset: %w", err)
		}
		st.View.Asset = &state.Asset{
			ID:        ins.Asset,
			Name:      ins.Asset,
			Url:       asset.URL,
			MediaType: asset.MediaType,
		}
	}

	// TODO: grid settings and layers come from the battle map format once it exists.
	st.Grid.Show = ins.View == scenes.ViewBattleMap

	return st, nil
}
