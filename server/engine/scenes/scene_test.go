package scenes_test

import (
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tavern() *scenes.Scene {
	return &scenes.Scene{
		ID:   "tavern",
		Name: "The Prancing Pony",
		Entities: []scenes.Entity{
			{ID: "barkeep", Kind: scenes.EntityNPC, Name: "Barliman"},
			{ID: "spy", Kind: scenes.EntityEnemy, Name: "Spy", Hidden: true},
		},
		Instructions: scenes.Instructions{
			Screens: map[string]scenes.ScreenInstruction{
				"table": {View: scenes.ViewBattleMap, Asset: "tavern.dd2vtt"},
				"wall":  {View: scenes.ViewScenery, Asset: "tavern-day.png"},
			},
			Lights: map[string]scenes.LightInstruction{
				"ambient": {On: true, Color: "#ffcc88", Brightness: 0.8},
			},
			Music: map[string]scenes.MusicInstruction{
				"music": {Action: scenes.MusicPlay, Source: "spotify:playlist:tavern", Volume: 0.4},
			},
		},
		Variants: []scenes.Variant{
			{
				ID:   "night",
				Tags: map[string]string{"time_of_day": "night"},
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{
						"wall": {View: scenes.ViewScenery, Asset: "tavern-night.png"},
					},
					Lights: map[string]scenes.LightInstruction{
						"ambient": {On: true, Color: "#ff8800", Brightness: 0.3, Effect: "candle"},
					},
				},
			},
		},
	}
}

var _ = Describe("Scene", func() {
	It("resolves base instructions without a variant", func() {
		ins, err := tavern().Resolve("")
		Expect(err).ToNot(HaveOccurred())
		Expect(ins.Screens["wall"].Asset).To(Equal("tavern-day.png"))
		Expect(ins.Lights["ambient"].Brightness).To(Equal(0.8))
	})

	It("applies variant overrides per target", func() {
		ins, err := tavern().Resolve("night")
		Expect(err).ToNot(HaveOccurred())
		Expect(ins.Screens["wall"].Asset).To(Equal("tavern-night.png"))
		Expect(ins.Screens["table"].Asset).To(Equal("tavern.dd2vtt"), "untouched targets are kept")
		Expect(ins.Lights["ambient"].Effect).To(Equal("candle"))
		Expect(ins.Music).To(HaveKey("music"))
	})

	It("does not modify the base instructions when resolving", func() {
		s := tavern()
		_, err := s.Resolve("night")
		Expect(err).ToNot(HaveOccurred())
		Expect(s.Instructions.Screens["wall"].Asset).To(Equal("tavern-day.png"))
	})

	It("uses the default variant when none is requested", func() {
		s := tavern()
		s.DefaultVariant = "night"
		ins, err := s.Resolve("")
		Expect(err).ToNot(HaveOccurred())
		Expect(ins.Screens["wall"].Asset).To(Equal("tavern-night.png"))
	})

	It("returns an error for unknown variants", func() {
		_, err := tavern().Resolve("rain")
		Expect(err).To(MatchError(scenes.ErrVariantNotFound))
	})

	It("filters hidden entities", func() {
		Expect(tavern().VisibleEntities()).To(HaveLen(1))
	})
})

var _ = Describe("Plan", func() {
	It("sends everything when nothing is active", func() {
		t, err := scenes.Plan(nil, "", tavern(), "")
		Expect(err).ToNot(HaveOccurred())
		Expect(t.From.IsZero()).To(BeTrue())
		Expect(t.To).To(Equal(scenes.State{SceneID: "tavern"}))
		Expect(t.Changes.All()).To(HaveLen(4))
		Expect(t.Released).To(BeEmpty())
	})

	It("only sends changed instructions on a variant transition", func() {
		s := tavern()
		t, err := scenes.Plan(s, "", s, "night")
		Expect(err).ToNot(HaveOccurred())
		Expect(t.To.VariantID).To(Equal("night"))
		Expect(t.Changes.Screens).To(HaveLen(1))
		Expect(t.Changes.Screens).To(HaveKey("wall"))
		Expect(t.Changes.Lights).To(HaveKey("ambient"))
		Expect(t.Changes.Music).To(BeEmpty())
	})

	It("releases targets the next scene does not instruct", func() {
		road := &scenes.Scene{
			ID: "road",
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{
					"table": {View: scenes.ViewBattleMap, Asset: "road.dd2vtt"},
				},
			},
		}
		t, err := scenes.Plan(tavern(), "", road, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(t.Changes.All()).To(HaveLen(1))
		Expect(t.Released).To(ConsistOf(
			scenes.TargetRef{Kind: devices.KindScreen, Target: "wall"},
			scenes.TargetRef{Kind: devices.KindLight, Target: "ambient"},
			scenes.TargetRef{Kind: devices.KindMusic, Target: "music"},
		))
	})
})
