package scenes_test

import (
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func load(cfg *scenes.Config, state *scenes.State) *scenes.Scene {
	GinkgoHelper()
	s, err := scenes.Load(cfg, state)
	Expect(err).ToNot(HaveOccurred())

	return s
}

func instructions(s *scenes.Scene) scenes.Instructions {
	GinkgoHelper()
	ins, err := s.Instructions()
	Expect(err).ToNot(HaveOccurred())

	return ins
}

var _ = Describe("Scene", func() {
	Describe("Load", func() {
		It("creates a fresh state from the config", func() {
			cfg := tavern()
			cfg.DefaultVariant = "night"
			s := load(cfg, nil)
			Expect(s.State.SceneID).To(Equal("tavern"))
			Expect(s.State.VariantID).To(Equal("night"))
			Expect(s.State.Entities).To(HaveKeyWithValue("spy", scenes.EntityState{Hidden: true}))
		})

		It("restores a saved state", func() {
			state := &scenes.State{
				SceneID:   "tavern",
				VariantID: "night",
				Entities:  map[string]scenes.EntityState{"spy": {Hidden: false}},
			}
			s := load(tavern(), state)
			Expect(s.Ref()).To(Equal(scenes.Ref{SceneID: "tavern", VariantID: "night"}))
			Expect(s.VisibleEntities()).To(HaveLen(2), "spy was revealed, barkeep is new in the config")
		})

		It("falls back to the default variant when the saved one was removed from the config", func() {
			s := load(tavern(), &scenes.State{SceneID: "tavern", VariantID: "deleted"})
			Expect(s.State.VariantID).To(BeEmpty())
		})

		It("rejects a state of another scene", func() {
			_, err := scenes.Load(tavern(), &scenes.State{SceneID: "road"})
			Expect(err).To(MatchError(scenes.ErrStateMismatch))
		})

		It("never modifies the config", func() {
			cfg := tavern()
			s := load(cfg, nil)
			Expect(s.State.SetHidden("spy", false)).To(Succeed())
			Expect(s.SetVariant("night")).To(Succeed())
			s.Adjust(scenes.Instructions{Lights: map[string]scenes.LightInstruction{"ambient": {On: false}}})

			Expect(cfg).To(Equal(tavern()))
		})
	})

	Describe("entities", func() {
		It("tracks hidden, removed and spawned entities in the state", func() {
			s := load(tavern(), nil)
			Expect(s.VisibleEntities()).To(HaveLen(1))

			Expect(s.State.SetHidden("spy", false)).To(Succeed())
			Expect(s.State.Remove("barkeep")).To(Succeed())
			Expect(s.State.Spawn(scenes.Entity{ID: "wolf", Kind: scenes.EntityEnemy, Name: "Wolf"})).To(Succeed())

			visible := s.VisibleEntities()
			ids := make([]string, 0, len(visible))
			for _, e := range visible {
				ids = append(ids, e.ID)
			}
			Expect(ids).To(Equal([]string{"spy", "wolf"}))
		})

		It("rejects unknown and duplicate entities", func() {
			s := load(tavern(), nil)
			Expect(s.State.SetHidden("ghost", true)).To(MatchError(scenes.ErrEntityNotFound))
			Expect(s.State.Spawn(scenes.Entity{ID: "spy"})).To(MatchError(scenes.ErrEntityExists))
		})
	})

	Describe("instructions", func() {
		It("applies live overrides on top of the active variant", func() {
			s := load(tavern(), nil)
			s.Adjust(scenes.Instructions{Music: map[string]scenes.MusicInstruction{"music": {Action: scenes.MusicPause}}})

			ins := instructions(s)
			Expect(ins.Music["music"].Action).To(Equal(scenes.MusicPause))
			Expect(ins.Screens["wall"].Asset).To(Equal("tavern-day.png"))
		})

		It("clears live overrides when switching variant", func() {
			s := load(tavern(), nil)
			s.Adjust(scenes.Instructions{Music: map[string]scenes.MusicInstruction{"music": {Action: scenes.MusicPause}}})
			Expect(s.SetVariant("night")).To(Succeed())

			Expect(instructions(s).Music["music"].Action).To(Equal(scenes.MusicPlay))
		})
	})
})

var _ = Describe("Transition", func() {
	It("sends everything when nothing is active", func() {
		s := load(tavern(), nil)
		t := scenes.NewTransition(scenes.Ref{}, scenes.Instructions{}, s.Ref(), instructions(s))
		Expect(t.From.IsZero()).To(BeTrue())
		Expect(t.Changes.All()).To(HaveLen(4))
		Expect(t.Released).To(BeEmpty())
	})

	It("only sends changed instructions on a variant transition", func() {
		s := load(tavern(), nil)
		from, prev := s.Ref(), instructions(s)
		Expect(s.SetVariant("night")).To(Succeed())

		t := scenes.NewTransition(from, prev, s.Ref(), instructions(s))
		Expect(t.To.VariantID).To(Equal("night"))
		Expect(t.Changes.Screens).To(HaveLen(1))
		Expect(t.Changes.Screens).To(HaveKey("wall"))
		Expect(t.Changes.Lights).To(HaveKey("ambient"))
		Expect(t.Changes.Music).To(BeEmpty())
	})

	It("re-instructs a screen when a variant only adds overlays", func() {
		s := load(tavern(), nil)
		from, prev := s.Ref(), instructions(s)
		Expect(s.SetVariant("combat")).To(Succeed())

		t := scenes.NewTransition(from, prev, s.Ref(), instructions(s))
		Expect(t.Changes.All()).To(HaveLen(1))
		Expect(t.Changes.Screens["wall"].Overlays).To(HaveLen(2))
	})

	It("releases targets the next scene does not instruct", func() {
		prev := load(tavern(), nil)
		next := load(&scenes.Config{
			ID: "road",
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{
					"table": {View: scenes.ViewBattleMap, Asset: "road-map"},
				},
			},
		}, nil)

		t := scenes.NewTransition(prev.Ref(), instructions(prev), next.Ref(), instructions(next))
		Expect(t.Changes.All()).To(HaveLen(1))
		Expect(t.Released).To(ConsistOf(
			scenes.TargetRef{Kind: devices.KindScreen, Target: "wall"},
			scenes.TargetRef{Kind: devices.KindLight, Target: "ambient"},
			scenes.TargetRef{Kind: devices.KindMusic, Target: "music"},
		))
	})
})
