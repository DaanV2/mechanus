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
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tavernMap = `{
	"version": 0,
	"background_color": "#101010",
	"grid": {"cell_size": 50, "columns": 8, "rows": 6, "x_offset": 2, "y_offset": 3, "line_color": "#ffffff"},
	"layers": [
		{"id": "floor", "kind": "image", "asset": "images/floor.png", "opacity": 0.5},
		{"id": "rug", "kind": "image", "asset": "shared:/rug.png", "position": {"x": 1, "y": 2}, "size": {"x": 3, "y": 4}},
		{
			"id": "tokens",
			"kind": "entities",
			"instance_groups": [{"id": "barrels", "asset": "shared:/barrel.png", "instances": [{"x": 1, "y": 1}, {"x": 2, "y": 1, "rotation": 45}]}],
			"entities": [
				{"entity": "barkeep", "asset": "tokens/barkeep.png", "position": {"x": 4, "y": 5}, "size": {"x": 2, "y": 2}, "rotation": 90},
				{"entity": "hero", "position": {"x": 1, "y": 3}},
				{"entity": "spy", "asset": "tokens/spy.png", "position": {"x": 6, "y": 1}}
			]
		}
	]
}`

var _ = Describe("Driver", func() {
	var (
		manager *screens.ScreenManager
		driver  *screen.Driver
		ref     = scenes.Ref{SceneID: "tavern"}
	)

	present := []scenes.Entity{
		{ID: "barkeep", Kind: scenes.EntityNPC, Name: "Barkeep"},
		{ID: "hero", Kind: scenes.EntityPlayer, Name: "Hero"},
	}

	BeforeEach(func() {
		manager = screens.NewScreenManager()
		driver = screen.NewDriver(manager, assets.NewStore(fstest.MapFS{
			"scenes/tavern/maps/tavern.map.json": {Data: []byte(tavernMap)},
			"scenes/tavern/maps/broken.map.json": {Data: []byte(`{"version": 7}`)},
			"scenes/tavern/maps/lost.map.json":   {Data: []byte(`{"grid": {"cell_size": 1, "columns": 1, "rows": 1}, "layers": [{"id": "a", "kind": "image", "asset": "gone.png"}]}`)},
			"scenes/tavern/images/floor.png":     {Data: []byte("floor")},
			"scenes/tavern/tokens/barkeep.png":   {Data: []byte("barkeep")},
			"scenes/tavern/tokens/spy.png":       {Data: []byte("spy")},
			"scenes/tavern/wall.png":             {Data: []byte("wall")},
			"shared/rug.png":                     {Data: []byte("rug")},
			"shared/barrel.png":                  {Data: []byte("barrel")},
		}))
	})

	battleMap := func() *state.ScreenState {
		GinkgoHelper()
		st, err := driver.State(ref, &scenes.ScreenInstruction{
			View:     scenes.ViewBattleMap,
			Asset:    "maps/tavern.map.json",
			Overlays: []scenes.Overlay{{Kind: scenes.OverlayEntities}},
			Entities: present,
		})
		Expect(err).ToNot(HaveOccurred())

		return st
	}

	It("sends the grid settings and background of a battle map", func() {
		st := battleMap()

		Expect(st.Environment.BackgroundColor).To(Equal("#101010"))
		Expect(st.Grid.Show).To(BeTrue())
		Expect(st.Grid.CellSize).To(Equal(50.0))
		Expect(st.Grid.Columns).To(Equal(8))
		Expect(st.Grid.Rows).To(Equal(6))
		Expect(st.Grid.XOffset).To(Equal(2))
		Expect(st.Grid.YOffset).To(Equal(3))
		Expect(st.Grid.LineColor).To(Equal("#ffffff"))
		Expect(st.View.Kind).To(Equal(scenes.ViewBattleMap))
		Expect(st.View.Asset).To(BeNil(), "the map file itself is not for the client")
		Expect(st.View.Overlays).To(HaveLen(1))
	})

	It("sends the image layers in order, with resolved assets", func() {
		st := battleMap()

		Expect(st.Layers).To(HaveLen(3))
		floor, rug := st.Layers[0], st.Layers[1]
		Expect(floor.ID).To(Equal("floor"))
		Expect(floor.Kind).To(Equal(scenes.MapLayerImage))
		Expect(floor.Asset.Url).To(HavePrefix("/assets/scenes/tavern/images/floor.png?v="))
		Expect(floor.Opacity).To(Equal(0.5))
		Expect(floor.Size).To(BeNil())

		Expect(rug.ID).To(Equal("rug"))
		Expect(rug.Asset.Url).To(HavePrefix("/assets/shared/rug.png?v="))
		Expect(rug.Opacity).To(Equal(1.0), "defaults to opaque")
		Expect(rug.Position).To(Equal(vectors.Vector2D{X: 1, Y: 2}))
		Expect(*rug.Size).To(Equal(vectors.Vector2D{X: 3, Y: 4}))
	})

	It("sends the present entities with their placement", func() {
		tokens := battleMap().Layers[2]

		Expect(tokens.Entities).To(HaveLen(2))
		barkeep := tokens.Entities["barkeep"]
		Expect(barkeep.Name).To(Equal("Barkeep"))
		Expect(barkeep.Kind).To(Equal(scenes.EntityNPC))
		Expect(barkeep.Asset.Url).To(HavePrefix("/assets/scenes/tavern/tokens/barkeep.png?v="))
		Expect(barkeep.Position).To(Equal(vectors.Vector2D{X: 4, Y: 5}))
		Expect(barkeep.Size).To(Equal(vectors.Vector2D{X: 2, Y: 2}))
		Expect(barkeep.Rotation).To(Equal(90.0))

		hero := tokens.Entities["hero"]
		Expect(hero.Asset).To(BeNil(), "no asset is a placeholder token")
		Expect(hero.Size).To(Equal(vectors.Vector2D{X: 1, Y: 1}), "defaults to one cell")
	})

	It("does not send placed entities that are hidden or removed", func() {
		tokens := battleMap().Layers[2]

		Expect(tokens.Entities).ToNot(HaveKey("spy"))
	})

	It("sends instance groups", func() {
		groups := battleMap().Layers[2].InstanceGroups

		Expect(groups).To(HaveLen(1))
		Expect(groups[0].ID).To(Equal("barrels"))
		Expect(groups[0].Asset.Url).To(HavePrefix("/assets/shared/barrel.png?v="))
		Expect(groups[0].Size).To(Equal(vectors.Vector2D{X: 1, Y: 1}))
		Expect(groups[0].Instances).To(Equal([]scenes.Instance{{X: 1, Y: 1}, {X: 2, Y: 1, Rotation: 45}}))
	})

	It("converts all of it to the message sent to screens", func() {
		msg := battleMap().ToProto()

		Expect(msg.GetGrid().GetColumns()).To(Equal(int32(8)))
		Expect(msg.GetGrid().GetRows()).To(Equal(int32(6)))
		Expect(msg.GetLayer()).To(HaveLen(3))
		Expect(msg.GetLayer()[0].GetKind()).To(Equal(screensv1.LayerKind_LAYER_KIND_IMAGE))
		Expect(msg.GetLayer()[0].GetSize()).To(BeNil())
		Expect(msg.GetLayer()[1].GetSize().GetX()).To(Equal(3.0))

		tokens := msg.GetLayer()[2]
		Expect(tokens.GetKind()).To(Equal(screensv1.LayerKind_LAYER_KIND_ENTITIES))
		Expect(tokens.GetEntity()).To(HaveLen(2))
		Expect(tokens.GetEntity()[0].GetId()).To(Equal("barkeep"))
		Expect(tokens.GetInstanceGroup()[0].GetInstance()[1].GetRotation()).To(Equal(45.0))
	})

	It("shows a plain image as a battle map under a default grid", func() {
		st, err := driver.State(ref, &scenes.ScreenInstruction{View: scenes.ViewBattleMap, Asset: "wall.png"})
		Expect(err).ToNot(HaveOccurred())

		Expect(st.View.Asset.Url).To(HavePrefix("/assets/scenes/tavern/wall.png?v="))
		Expect(st.Grid.Show).To(BeTrue())
		Expect(st.Grid.CellSize).To(Equal(float64(scenes.DefaultCellSize)))
		Expect(st.Grid.Columns).To(BeZero(), "the image decides the size")
		Expect(st.Layers).To(HaveLen(1))
		Expect(st.Layers[0].Kind).To(Equal(scenes.MapLayerImage))
		Expect(st.Layers[0].Asset.Url).To(Equal(st.View.Asset.Url))
	})

	It("does not send a battle map for other views", func() {
		st, err := driver.State(ref, &scenes.ScreenInstruction{View: scenes.ViewScenery, Asset: "wall.png"})
		Expect(err).ToNot(HaveOccurred())

		Expect(st.View.Asset.Url).To(HavePrefix("/assets/scenes/tavern/wall.png?v="))
		Expect(st.Grid.Show).To(BeFalse())
		Expect(st.Layers).To(BeEmpty())
	})

	DescribeTable("fails on a map it can't use",
		func(asset, message string) {
			_, err := driver.State(ref, &scenes.ScreenInstruction{View: scenes.ViewBattleMap, Asset: asset})
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("a missing file", "maps/gone.map.json", "maps/gone.map.json"),
		Entry("an invalid file", "maps/broken.map.json", "unsupported version 7"),
		Entry("a missing layer asset", "maps/lost.map.json", "gone.png"),
		Entry("a missing image", "gone.png", "gone.png"),
	)

	It("pushes the state to the device's screen", func(ctx SpecContext) {
		err := driver.Apply(ctx, ref, &devices.Device{ID: "tv"}, scenes.ScreenInstruction{
			View:     scenes.ViewBattleMap,
			Asset:    "maps/tavern.map.json",
			Entities: present,
		})
		Expect(err).ToNot(HaveOccurred())

		h, ok := manager.Get("tv")
		Expect(ok).To(BeTrue())
		Expect(h.State().GetLayer()).To(HaveLen(3))
		Expect(h.State().GetLayer()[2].GetEntity()).To(HaveLen(2))
	})

	It("ignores instructions of other kinds", func(ctx SpecContext) {
		err := driver.Apply(ctx, ref, &devices.Device{ID: "tv"}, scenes.LightInstruction{On: true})
		Expect(err).ToNot(HaveOccurred())

		_, ok := manager.Get("tv")
		Expect(ok).To(BeFalse())
	})
})
