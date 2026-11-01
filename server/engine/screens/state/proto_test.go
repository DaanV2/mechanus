package state_test

import (
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/engine/screens/state"
	"github.com/DaanV2/mechanus/server/pkg/math/vectors"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScreenState.ToProto", func() {
	It("converts the default state", func() {
		msg := state.NewScreenState().ToProto()

		Expect(msg.GetEnvironment().GetBackgroundColor()).To(Equal("#000000"))
		Expect(msg.GetEnvironment().GetCurrentTime()).NotTo(BeNil())
		Expect(msg.GetGrid().GetXRatio()).To(Equal(int32(1)))
		Expect(msg.GetView().GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BLANK))
		Expect(msg.GetCamera().GetZoom()).To(Equal(1.0))
		Expect(msg.GetLayer()).To(BeEmpty())
		Expect(msg.GetSplashScreen()).To(BeNil())
	})

	It("converts the view, overlays, layers and entities", func() {
		s := state.NewScreenState()
		s.View = &state.View{
			Kind:   scenes.ViewBattleMap,
			Asset:  &state.Asset{ID: "map", Url: "/assets/tavern.png", MediaType: "image/png"},
			Params: map[string]string{"fog": "true"},
			Overlays: []scenes.Overlay{
				{Kind: scenes.OverlayInitiative},
				{Kind: scenes.OverlayCustom, Component: "clock"},
			},
		}
		s.Camera = &state.Camera{Center: vectors.Vector2D{X: 4, Y: 5}, Zoom: 2}
		s.Layers = []*state.Layer{{
			ID: "tokens",
			Entities: map[string]*state.Entity{
				"b": {ID: "b", Name: "Spy", Kind: scenes.EntityEnemy, Hidden: true},
				"a": {
					ID:       "a",
					Name:     "Barkeep",
					Kind:     scenes.EntityNPC,
					Position: vectors.Vector2D{X: 1, Y: 2},
					Size:     vectors.Vector2D{X: 1, Y: 1},
					Rotation: 90,
				},
			},
		}}
		s.SplashScreen = &state.SplashScreen{Show: true, Title: "MECHANUS", BackgroundColor: "#111111"}

		msg := s.ToProto()

		view := msg.GetView()
		Expect(view.GetKind()).To(Equal(screensv1.ViewKind_VIEW_KIND_BATTLEMAP))
		Expect(view.GetAsset().GetUrl()).To(Equal("/assets/tavern.png"))
		Expect(view.GetAsset().GetMediaType()).To(Equal("image/png"))
		Expect(view.GetParams()).To(HaveKeyWithValue("fog", "true"))
		Expect(view.GetComponent()).To(BeEmpty())
		Expect(view.GetOverlay()).To(HaveLen(2))
		Expect(view.GetOverlay()[0].GetKind()).To(Equal(screensv1.OverlayKind_OVERLAY_KIND_INITIATIVE))
		Expect(view.GetOverlay()[1].GetComponent()).To(Equal("clock"))

		Expect(msg.GetCamera().GetCenter().GetX()).To(Equal(4.0))
		Expect(msg.GetCamera().GetZoom()).To(Equal(2.0))

		Expect(msg.GetLayer()).To(HaveLen(1))
		layer := msg.GetLayer()[0]
		Expect(layer.GetId()).To(Equal("tokens"))
		Expect(layer.GetEntity()).To(HaveLen(2))

		barkeep, spy := layer.GetEntity()[0], layer.GetEntity()[1]
		Expect(barkeep.GetId()).To(Equal("a"), "entities are sorted by ID")
		Expect(barkeep.GetKind()).To(Equal(screensv1.EntityKind_ENTITY_KIND_NPC))
		Expect(barkeep.GetPosition().GetY()).To(Equal(2.0))
		Expect(barkeep.GetRotation()).To(Equal(90.0))
		Expect(barkeep.GetHidden()).To(BeFalse())
		Expect(spy.GetKind()).To(Equal(screensv1.EntityKind_ENTITY_KIND_ENEMY))
		Expect(spy.GetHidden()).To(BeTrue())

		Expect(msg.GetSplashScreen().GetTitle()).To(Equal("MECHANUS"))
		Expect(msg.GetSplashScreen().GetBackgroundHex()).To(Equal("#111111"))
	})

	It("maps every scene view, overlay and entity kind", func() {
		for _, k := range []scenes.ViewKind{
			scenes.ViewBattleMap, scenes.ViewScenery, scenes.ViewInfoBoard,
			scenes.ViewCustom, scenes.ViewSplash, scenes.ViewBlank,
		} {
			v := &state.View{Kind: k}
			Expect(v.ToProto().GetKind()).NotTo(Equal(screensv1.ViewKind_VIEW_KIND_UNSPECIFIED), string(k))
		}
		for _, k := range []scenes.OverlayKind{scenes.OverlayInitiative, scenes.OverlayEntities, scenes.OverlayCustom} {
			v := &state.View{Overlays: []scenes.Overlay{{Kind: k}}}
			Expect(v.ToProto().GetOverlay()[0].GetKind()).NotTo(Equal(screensv1.OverlayKind_OVERLAY_KIND_UNSPECIFIED), string(k))
		}
		for _, k := range []scenes.EntityKind{scenes.EntityPlayer, scenes.EntityNPC, scenes.EntityEnemy, scenes.EntityObject} {
			e := &state.Entity{Kind: k}
			Expect(e.ToProto().GetKind()).NotTo(Equal(screensv1.EntityKind_ENTITY_KIND_UNSPECIFIED), string(k))
		}
	})

	It("handles a nil state", func() {
		var s *state.ScreenState
		Expect(s.ToProto()).To(BeNil())
	})
})
