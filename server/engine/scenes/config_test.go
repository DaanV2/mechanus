package scenes_test

import (
	"github.com/DaanV2/mechanus/server/engine/scenes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tavern() *scenes.Config {
	return &scenes.Config{
		ID:   "tavern",
		Name: "The Prancing Pony",
		Entities: []scenes.Entity{
			{ID: "barkeep", Kind: scenes.EntityNPC, Name: "Barliman"},
			{ID: "spy", Kind: scenes.EntityEnemy, Name: "Spy", Hidden: true},
		},
		Instructions: scenes.Instructions{
			Screens: map[string]scenes.ScreenInstruction{
				"table": {View: scenes.ViewBattleMap, Asset: "tavern-map"},
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
			{
				ID: "combat",
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{
						"wall": {
							View:  scenes.ViewScenery,
							Asset: "tavern-day.png",
							Overlays: []scenes.Overlay{
								{Kind: scenes.OverlayInitiative},
								{Kind: scenes.OverlayEntities},
							},
						},
					},
				},
			},
		},
	}
}

var _ = Describe("Config", func() {
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
		Expect(ins.Screens["table"].Asset).To(Equal("tavern-map"), "untouched targets are kept")
		Expect(ins.Lights["ambient"].Effect).To(Equal("candle"))
		Expect(ins.Music).To(HaveKey("music"))
	})

	It("does not modify the base instructions when resolving", func() {
		c := tavern()
		_, err := c.Resolve("night")
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Instructions.Screens["wall"].Asset).To(Equal("tavern-day.png"))
	})

	It("uses the default variant when none is requested", func() {
		c := tavern()
		c.DefaultVariant = "night"
		ins, err := c.Resolve("")
		Expect(err).ToNot(HaveOccurred())
		Expect(ins.Screens["wall"].Asset).To(Equal("tavern-night.png"))
	})

	It("returns an error for unknown variants", func() {
		_, err := tavern().Resolve("rain")
		Expect(err).To(MatchError(scenes.ErrVariantNotFound))
	})
})
