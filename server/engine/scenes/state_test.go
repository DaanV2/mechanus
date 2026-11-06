package scenes_test

import (
	"encoding/json"
	"math"

	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("State during play", func() {
	Describe("Move", func() {
		It("stores the position in the state, leaving the config's position the default", func() {
			s := load(tavern(), nil)
			_, moved := s.Position("barkeep")
			Expect(moved).To(BeFalse(), "never moved: the battle map decides")

			Expect(s.State.Move("barkeep", vectors.Vector2D{X: 3, Y: 4})).To(Succeed())

			pos, moved := s.Position("barkeep")
			Expect(moved).To(BeTrue())
			Expect(pos).To(Equal(vectors.Vector2D{X: 3, Y: 4}))
			Expect(s.Config).To(Equal(tavern()), "the config is never modified")
		})

		It("gives the battle map the positions of visible entities only", func() {
			s := load(tavern(), nil)
			Expect(s.State.Move("barkeep", vectors.Vector2D{X: 1, Y: 2})).To(Succeed())
			Expect(s.State.Move("spy", vectors.Vector2D{X: 5, Y: 5})).To(Succeed())

			ins := instructions(s)
			Expect(ins.Screens["table"].Positions).To(Equal(map[string]vectors.Vector2D{"barkeep": {X: 1, Y: 2}}))
			Expect(ins.Screens["wall"].Positions).To(BeNil(), "not a battle map")
		})

		It("re-instructs only the battle map", func() {
			s := load(tavern(), nil)
			prev := instructions(s)
			Expect(s.State.Move("barkeep", vectors.Vector2D{X: 1, Y: 2})).To(Succeed())

			t := scenes.NewTransition(s.Ref(), prev, s.Ref(), instructions(s))
			Expect(t.Changes.All()).To(HaveLen(1))
			Expect(t.Changes.Screens).To(HaveKey("table"))
		})

		It("keeps a moved entity's position when it is hidden and revealed", func() {
			s := load(tavern(), nil)
			Expect(s.State.Move("barkeep", vectors.Vector2D{X: 1, Y: 2})).To(Succeed())
			Expect(s.State.SetHidden("barkeep", true)).To(Succeed())
			Expect(s.State.SetHidden("barkeep", false)).To(Succeed())

			Expect(instructions(s).Screens["table"].Positions).To(HaveKeyWithValue("barkeep", vectors.Vector2D{X: 1, Y: 2}))
		})

		DescribeTable("refuses invalid moves",
			func(entityID string, pos vectors.Vector2D, expected error) {
				s := load(tavern(), nil)
				Expect(s.State.Remove("barkeep")).To(Succeed())
				Expect(s.State.Move(entityID, pos)).To(MatchError(expected))
			},
			Entry("an unknown entity", "ghost", vectors.Vector2D{}, scenes.ErrEntityNotFound),
			Entry("a removed entity", "barkeep", vectors.Vector2D{}, scenes.ErrEntityNotFound),
			Entry("not a number", "spy", vectors.Vector2D{X: math.NaN()}, scenes.ErrInvalidPosition),
			Entry("infinity", "spy", vectors.Vector2D{Y: math.Inf(1)}, scenes.ErrInvalidPosition),
			Entry("far off the map", "spy", vectors.Vector2D{X: scenes.MaxCoordinate + 1}, scenes.ErrInvalidPosition),
		)

		It("survives saving and restoring the state", func() {
			s := load(tavern(), nil)
			Expect(s.State.Move("spy", vectors.Vector2D{X: 7, Y: 8})).To(Succeed())
			Expect(s.State.SetCamera("table", scenes.Camera{Center: vectors.Vector2D{X: 2, Y: 3}, Zoom: 2})).To(Succeed())

			data, err := json.Marshal(s.State)
			Expect(err).ToNot(HaveOccurred())
			var restored scenes.State
			Expect(json.Unmarshal(data, &restored)).To(Succeed())

			r := load(tavern(), &restored)
			pos, moved := r.Position("spy")
			Expect(moved).To(BeTrue())
			Expect(pos).To(Equal(vectors.Vector2D{X: 7, Y: 8}))
			Expect(r.State.Cameras).To(HaveKeyWithValue("table", scenes.Camera{Center: vectors.Vector2D{X: 2, Y: 3}, Zoom: 2}))
		})

		It("only stores the entities that moved", func() {
			s := load(tavern(), nil)
			Expect(s.State.Move("spy", vectors.Vector2D{X: 7, Y: 8})).To(Succeed())

			data, err := json.Marshal(s.State.Entities)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(MatchJSON(`{
				"barkeep": {"hidden": false, "removed": false},
				"spy": {"hidden": true, "removed": false, "position": {"x": 7, "y": 8}}
			}`))
		})
	})

	Describe("SetCamera", func() {
		It("gives the battle map of the target its camera", func() {
			s := load(tavern(), nil)
			Expect(instructions(s).Screens["table"].Camera).To(BeNil(), "the whole map by default")

			camera := scenes.Camera{Center: vectors.Vector2D{X: 4, Y: 2}, Zoom: 1.5}
			Expect(s.State.SetCamera("table", camera)).To(Succeed())

			Expect(instructions(s).Screens["table"].Camera).To(Equal(&camera))
		})

		DescribeTable("refuses invalid cameras",
			func(camera scenes.Camera) {
				s := load(tavern(), nil)
				Expect(s.State.SetCamera("table", camera)).To(MatchError(scenes.ErrInvalidCamera))
				Expect(s.State.Cameras).To(BeEmpty())
			},
			Entry("zoomed out too far", scenes.Camera{Zoom: scenes.MinZoom / 2}),
			Entry("zoomed in too far", scenes.Camera{Zoom: scenes.MaxZoom * 2}),
			Entry("no zoom", scenes.Camera{}),
			Entry("a zoom that is not a number", scenes.Camera{Zoom: math.NaN()}),
			Entry("a center that is not a number", scenes.Camera{Center: vectors.Vector2D{X: math.NaN()}, Zoom: 1}),
		)
	})

	Describe("BattleMapTargets", func() {
		It("lists the screen targets that show a battle map", func() {
			targets, err := load(tavern(), nil).BattleMapTargets()
			Expect(err).ToNot(HaveOccurred())
			Expect(targets).To(Equal([]string{"table"}))
		})
	})
})
