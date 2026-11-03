package scenes_test

import (
	"os"
	"strings"

	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BattleMap", func() {
	const valid = `{
		"version": 0,
		"name": "Cave",
		"background_color": "#101010",
		"grid": {"cell_size": 70, "columns": 16, "rows": 9, "x_offset": 3, "y_offset": 4, "line_color": "#ffffff"},
		"layers": [
			{"id": "floor", "kind": "image", "asset": "maps/floor.webp", "opacity": 0.5},
			{"id": "roof", "kind": "image", "asset": "shared:/roof.png", "position": {"x": 2, "y": 3}, "size": {"x": 4, "y": 5}},
			{
				"id": "tokens",
				"kind": "entities",
				"instance_groups": [{"id": "trees", "asset": "tree.png", "instances": [{"x": 1, "y": 2}, {"x": 3, "y": 4, "rotation": 90}]}],
				"entities": [{"entity": "goblin", "asset": "goblin.png", "position": {"x": 5, "y": 6}, "size": {"x": 2, "y": 2}, "rotation": 45}]
			}
		]
	}`

	It("reads a battle map and keeps its layers in order", func() {
		m, err := scenes.ParseBattleMap(strings.NewReader(valid))
		Expect(err).ToNot(HaveOccurred())

		Expect(m.Name).To(Equal("Cave"))
		Expect(m.BackgroundColor).To(Equal("#101010"))
		Expect(m.Grid).To(Equal(scenes.MapGrid{CellSize: 70, Columns: 16, Rows: 9, XOffset: 3, YOffset: 4, LineColor: "#ffffff"}))

		Expect(m.Layers).To(HaveLen(3))
		Expect(m.Layers[0].ID).To(Equal("floor"))
		Expect(*m.Layers[0].Opacity).To(Equal(0.5))
		Expect(m.Layers[1].Kind).To(Equal(scenes.MapLayerImage))
		Expect(m.Layers[1].Position).To(Equal(vectors.Vector2D{X: 2, Y: 3}))
		Expect(*m.Layers[1].Size).To(Equal(vectors.Vector2D{X: 4, Y: 5}))

		tokens := m.Layers[2]
		Expect(tokens.Kind).To(Equal(scenes.MapLayerEntities))
		Expect(tokens.Entities).To(HaveLen(1))
		Expect(tokens.Entities[0].Entity).To(Equal("goblin"))
		Expect(tokens.Entities[0].Rotation).To(Equal(45.0))
		Expect(tokens.InstanceGroups).To(HaveLen(1))
		Expect(tokens.InstanceGroups[0].Instances).To(Equal([]scenes.Instance{{X: 1, Y: 2}, {X: 3, Y: 4, Rotation: 90}}))
	})

	DescribeTable("rejects invalid maps",
		func(content, message string) {
			_, err := scenes.ParseBattleMap(strings.NewReader(content))
			Expect(err).To(MatchError(scenes.ErrInvalidBattleMap))
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("not JSON", `nope`, "invalid character"),
		Entry("an unknown field", grid(`"layrs": []`), "unknown field"),
		Entry("a newer version", `{"version": 1, "grid": {"cell_size": 70, "columns": 1, "rows": 1}}`, "unsupported version 1"),
		Entry("no cell size", `{"grid": {"columns": 1, "rows": 1}}`, "cell_size must be positive"),
		Entry("no columns", `{"grid": {"cell_size": 70, "rows": 1}}`, "columns and rows must be positive"),
		Entry("a bad line color", `{"grid": {"cell_size": 70, "columns": 1, "rows": 1, "line_color": "red"}}`, "line_color"),
		Entry("a bad background color", grid(`"background_color": "#12"`), "background_color"),
		Entry("a layer without id", layers(`{"kind": "image", "asset": "a.png"}`), "empty id"),
		Entry("duplicate layer ids", layers(`{"id": "a", "kind": "image", "asset": "a.png"}, {"id": "a", "kind": "image", "asset": "a.png"}`), "duplicate id"),
		Entry("an unknown layer kind", layers(`{"id": "a", "kind": "walls"}`), `unknown kind "walls"`),
		Entry("an image layer without asset", layers(`{"id": "a", "kind": "image"}`), "without asset"),
		Entry("an opacity out of range", layers(`{"id": "a", "kind": "image", "asset": "a.png", "opacity": 2}`), "opacity"),
		Entry("a negative size", layers(`{"id": "a", "kind": "image", "asset": "a.png", "size": {"x": -1, "y": 1}}`), "size must be positive"),
		Entry("an image layer with entities", layers(`{"id": "a", "kind": "image", "asset": "a.png", "entities": [{"entity": "e"}]}`), "image layer with entities"),
		Entry("an entities layer with an image", layers(`{"id": "a", "kind": "entities", "asset": "a.png"}`), "entities layer with image fields"),
		Entry("a placement without entity", layers(`{"id": "a", "kind": "entities", "entities": [{"position": {"x": 1, "y": 1}}]}`), "empty entity id"),
		Entry("an entity placed twice", layers(`{"id": "a", "kind": "entities", "entities": [{"entity": "e"}]}, {"id": "b", "kind": "entities", "entities": [{"entity": "e"}]}`), "placed more than once"),
		Entry("an instance group without id", layers(`{"id": "a", "kind": "entities", "instance_groups": [{"asset": "a.png"}]}`), "empty id"),
		Entry("an instance group without asset", layers(`{"id": "a", "kind": "entities", "instance_groups": [{"id": "g"}]}`), "no asset"),
		Entry("duplicate instance groups", layers(`{"id": "a", "kind": "entities", "instance_groups": [{"id": "g", "asset": "a.png"}, {"id": "g", "asset": "a.png"}]}`), "duplicate id"),
	)

	It("recognizes battle map files", func() {
		Expect(scenes.IsBattleMapFile("maps/cave.map.json")).To(BeTrue())
		Expect(scenes.IsBattleMapFile("shared:/maps/cave.map.json")).To(BeTrue())
		Expect(scenes.IsBattleMapFile("maps/cave.webp")).To(BeFalse())
	})

	It("creates a default map for a plain image", func() {
		m := scenes.ImageBattleMap("maps/cave.webp")

		Expect(m.Grid.CellSize).To(Equal(float64(scenes.DefaultCellSize)))
		Expect(m.Layers).To(HaveLen(1))
		Expect(m.Layers[0].Asset).To(Equal("maps/cave.webp"))
	})

	DescribeTable("the example scene's maps are valid",
		func(file string) {
			f, err := os.DirFS("../../../docs/examples/battlemap/scenes/example-tavern/maps").Open(file)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = f.Close() }()

			_, err = scenes.ParseBattleMap(f)
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("day", "tavern.map.json"),
		Entry("night", "tavern-night.map.json"),
	)
})

var _ = Describe("Scene instructions for a battle map", func() {
	newScene := func() *scenes.Scene {
		GinkgoHelper()
		s, err := scenes.Load(&scenes.Config{
			ID: "cave",
			Entities: []scenes.Entity{
				{ID: "goblin", Kind: scenes.EntityEnemy, Name: "Goblin"},
				{ID: "spy", Kind: scenes.EntityEnemy, Name: "Spy", Hidden: true},
			},
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{
					"table": {View: scenes.ViewBattleMap, Asset: "maps/cave.map.json"},
					"wall":  {View: scenes.ViewScenery, Asset: "cave.webp"},
				},
			},
		}, nil)
		Expect(err).ToNot(HaveOccurred())

		return s
	}
	instructions := func(s *scenes.Scene) scenes.Instructions {
		GinkgoHelper()
		ins, err := s.Instructions()
		Expect(err).ToNot(HaveOccurred())

		return ins
	}
	entityIDs := func(ins scenes.ScreenInstruction) []string {
		ids := make([]string, 0, len(ins.Entities))
		for _, e := range ins.Entities {
			ids = append(ids, e.ID)
		}

		return ids
	}

	It("gives the battle map the visible entities, and other views none", func() {
		ins := instructions(newScene())

		Expect(entityIDs(ins.Screens["table"])).To(ConsistOf("goblin"))
		Expect(ins.Screens["wall"].Entities).To(BeEmpty())
	})

	It("re-instructs only the battle map when an entity is revealed", func() {
		s := newScene()
		prev := instructions(s)

		Expect(s.State.SetHidden("spy", false)).To(Succeed())
		t := scenes.NewTransition(s.Ref(), prev, s.Ref(), instructions(s))

		Expect(t.Changes.Screens).To(HaveLen(1))
		Expect(entityIDs(t.Changes.Screens["table"])).To(ConsistOf("goblin", "spy"))
	})

	It("does not modify the configuration or store the entities", func() {
		s := newScene()
		instructions(s)

		Expect(s.Config.Instructions.Screens["table"].Entities).To(BeNil())
	})
})

func grid(field string) string {
	return `{"grid": {"cell_size": 70, "columns": 1, "rows": 1}, ` + field + `}`
}

func layers(content string) string {
	return grid(`"layers": [` + content + `]`)
}
