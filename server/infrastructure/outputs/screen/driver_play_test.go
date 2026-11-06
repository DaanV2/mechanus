package screen_test

import (
	"testing/fstest"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/screen"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Driver during play", func() {
	var (
		manager *screens.ScreenManager
		driver  *screen.Driver
		ref     = scenes.Ref{SceneID: "tavern"}
		tv      = &devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen}
	)

	BeforeEach(func() {
		manager = screens.NewScreenManager()
		driver = screen.NewDriver(manager, assets.NewStore(fstest.MapFS{
			"scenes/tavern/maps/tavern.map.json": {Data: []byte(tavernMap)},
			"scenes/tavern/images/floor.png":     {Data: []byte("floor")},
			"scenes/tavern/tokens/barkeep.png":   {Data: []byte("barkeep")},
			"scenes/tavern/tokens/spy.png":       {Data: []byte("spy")},
			"scenes/tavern/cave.png":             {Data: []byte("cave")},
			"shared/rug.png":                     {Data: []byte("rug")},
			"shared/barrel.png":                  {Data: []byte("barrel")},
		}))
	})

	stateFor := func(ins *scenes.ScreenInstruction) *state.ScreenState {
		GinkgoHelper()
		st, err := driver.State(ref, ins)
		Expect(err).ToNot(HaveOccurred())

		return st
	}

	entitiesOf := func(st *state.ScreenState, layerID string) map[string]*state.Entity {
		GinkgoHelper()
		for _, l := range st.Layers {
			if l.ID == layerID {
				return l.Entities
			}
		}
		Fail("no layer " + layerID)

		return nil
	}

	It("draws moved entities at their position from the state instead of the map's", func() {
		st := stateFor(&scenes.ScreenInstruction{
			View:      scenes.ViewBattleMap,
			Asset:     "maps/tavern.map.json",
			Entities:  []scenes.Entity{{ID: "hero", Kind: scenes.EntityPlayer}, {ID: "barkeep", Kind: scenes.EntityNPC}},
			Positions: map[string]vectors.Vector2D{"hero": {X: 7, Y: 2}},
		})

		tokens := entitiesOf(st, "tokens")
		Expect(tokens["hero"].Position).To(Equal(vectors.Vector2D{X: 7, Y: 2}))
		Expect(tokens["barkeep"].Position).To(Equal(vectors.Vector2D{X: 4, Y: 5}), "not moved: the map's position")
	})

	It("draws entities the map doesn't place on the first entity layer, once they have a position", func() {
		st := stateFor(&scenes.ScreenInstruction{
			View:  scenes.ViewBattleMap,
			Asset: "maps/tavern.map.json",
			Entities: []scenes.Entity{
				{ID: "wolf", Kind: scenes.EntityEnemy, Name: "Wolf"},
				{ID: "bat", Kind: scenes.EntityEnemy, Name: "Bat"},
			},
			Positions: map[string]vectors.Vector2D{"wolf": {X: 3, Y: 3}},
		})

		tokens := entitiesOf(st, "tokens")
		Expect(tokens).To(HaveKey("wolf"))
		Expect(tokens["wolf"]).To(Equal(&state.Entity{
			ID: "wolf", Name: "Wolf", Kind: scenes.EntityEnemy,
			Position: vectors.Vector2D{X: 3, Y: 3}, Size: vectors.Vector2D{X: 1, Y: 1},
		}))
		Expect(tokens).ToNot(HaveKey("bat"), "no position, so nowhere to draw it")
	})

	It("adds an entity layer to a map without one for entities with a position", func() {
		st := stateFor(&scenes.ScreenInstruction{
			View:      scenes.ViewBattleMap,
			Asset:     "cave.png",
			Entities:  []scenes.Entity{{ID: "wolf", Kind: scenes.EntityEnemy}},
			Positions: map[string]vectors.Vector2D{"wolf": {X: 1, Y: 1}},
		})

		Expect(st.Layers).To(HaveLen(2))
		Expect(entitiesOf(st, screen.PlayLayerID)).To(HaveKey("wolf"))
	})

	It("sends the camera from the state", func() {
		st := stateFor(&scenes.ScreenInstruction{
			View:   scenes.ViewBattleMap,
			Asset:  "maps/tavern.map.json",
			Camera: &scenes.Camera{Center: vectors.Vector2D{X: 2, Y: 1}, Zoom: 2},
		})

		Expect(st.Camera).To(Equal(&state.Camera{Center: vectors.Vector2D{X: 2, Y: 1}, Zoom: 2}))
	})

	It("moves a token on the device's screen without rebuilding its state", func(ctx SpecContext) {
		Expect(driver.Apply(ctx, ref, "table", tv, scenes.ScreenInstruction{
			View:     scenes.ViewBattleMap,
			Asset:    "maps/tavern.map.json",
			Entities: []scenes.Entity{{ID: "hero", Kind: scenes.EntityPlayer}},
		})).To(Succeed())

		Expect(driver.MoveEntity(ctx, ref, tv, "hero", vectors.Vector2D{X: 6, Y: 4})).To(Succeed())
		Expect(driver.MoveEntity(ctx, ref, tv, "spy", vectors.Vector2D{X: 6, Y: 4})).To(Succeed(), "not drawn: nothing to do")

		handler, _ := manager.Get("tv")
		for _, l := range handler.State().GetLayer() {
			for _, e := range l.GetEntity() {
				Expect(e.GetId()).To(Equal("hero"))
				Expect(e.GetPosition().GetX()).To(Equal(6.0))
			}
		}
	})

	It("moves the camera of the device's screen", func(ctx SpecContext) {
		Expect(driver.SetCamera(ctx, ref, tv, scenes.Camera{Center: vectors.Vector2D{X: 1, Y: 2}, Zoom: 4})).To(Succeed())

		handler, _ := manager.Get("tv")
		Expect(handler.State().GetCamera().GetZoom()).To(Equal(4.0))
	})
})
