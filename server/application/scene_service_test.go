package application_test

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type configStore map[string]*scenes.Config

func (c configStore) GetSceneConfig(_ context.Context, id string) (*scenes.Config, error) {
	if s, ok := c[id]; ok {
		return s, nil
	}

	return nil, errors.New("not found")
}

type stateStore struct {
	saved map[string]scenes.State // stored by value, as a database would
	saves int
}

func (s *stateStore) LoadSceneState(_ context.Context, id string) (*scenes.State, error) {
	if st, ok := s.saved[id]; ok {
		return &st, nil
	}

	return nil, scenes.ErrStateNotFound
}

func (s *stateStore) SaveSceneState(_ context.Context, st *scenes.State) error {
	s.saved[st.SceneID] = *st
	s.saves++

	return nil
}

type staticRoom struct{ room *devices.Room }

func (r staticRoom) CurrentRoom(context.Context) (*devices.Room, error) { return r.room, nil }

type applied struct {
	device      string
	instruction scenes.Instruction
}

type recordingDriver struct {
	integration string
	applied     []applied
}

func (d *recordingDriver) Integration() string { return d.integration }

func (d *recordingDriver) Apply(_ context.Context, device *devices.Device, ins scenes.Instruction) error {
	d.applied = append(d.applied, applied{device.ID, ins})

	return nil
}

var _ = Describe("SceneService", func() {
	var (
		ctx    context.Context
		room   *devices.Room
		states *stateStore
		screen *recordingDriver
		hue    *recordingDriver
		svc    *application.SceneService
	)

	BeforeEach(func() {
		ctx = context.Background()
		room = devices.NewRoom("room", "Living room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		room.AddDevice(&devices.Device{ID: "lamp", Kind: devices.KindLight, Integration: devices.IntegrationHue})
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(room.Bind("ambient", "lamp")).To(Succeed())

		configs := configStore{
			"cave": {
				ID:       "cave",
				Entities: []scenes.Entity{{ID: "troll", Kind: scenes.EntityEnemy, Hidden: true}},
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBattleMap, Asset: "cave-map"}},
					Lights:  map[string]scenes.LightInstruction{"ambient": {On: true, Brightness: 0.2}},
				},
				Variants: []scenes.Variant{{
					ID: "collapse",
					Instructions: scenes.Instructions{
						Lights: map[string]scenes.LightInstruction{"ambient": {On: true, Brightness: 0.2, Effect: "flicker"}},
					},
				}},
			},
			"road": {
				ID: "road",
				Instructions: scenes.Instructions{
					Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBattleMap, Asset: "road-map"}},
				},
			},
		}

		states = &stateStore{saved: map[string]scenes.State{}}
		screen = &recordingDriver{integration: devices.IntegrationScreen}
		hue = &recordingDriver{integration: devices.IntegrationHue}
		svc = application.NewSceneService(configs, states, staticRoom{room}, screen, hue)
	})

	It("instructs bound devices through their driver", func() {
		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.Active()).To(Equal(scenes.Ref{SceneID: "cave"}))
		Expect(screen.applied).To(HaveLen(1))
		Expect(screen.applied[0].device).To(Equal("tv"))
		Expect(hue.applied).To(HaveLen(1))
	})

	It("only re-instructs changed devices on a variant transition", func() {
		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.SetVariant(ctx, "collapse")
		Expect(err).ToNot(HaveOccurred())

		Expect(screen.applied).To(HaveLen(1))
		Expect(hue.applied).To(HaveLen(2))
		Expect(hue.applied[1].instruction).To(Equal(scenes.LightInstruction{On: true, Brightness: 0.2, Effect: "flicker"}))
	})

	It("dispatches live adjustments", func() {
		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.Adjust(ctx, scenes.Instructions{Lights: map[string]scenes.LightInstruction{"ambient": {On: false}}})
		Expect(err).ToNot(HaveOccurred())

		Expect(hue.applied).To(HaveLen(2))
		Expect(hue.applied[1].instruction).To(Equal(scenes.LightInstruction{On: false}))
	})

	It("saves state on load and exit, and restores it on the next load", func() {
		_, err := svc.Activate(ctx, "cave", "collapse")
		Expect(err).ToNot(HaveOccurred())
		Expect(states.saved).To(HaveKey("cave"), "saved on load")

		_, err = svc.Adjust(ctx, scenes.Instructions{Lights: map[string]scenes.LightInstruction{"ambient": {On: false}}})
		Expect(err).ToNot(HaveOccurred())

		_, err = svc.Activate(ctx, "road", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(states.saved["cave"].VariantID).To(Equal("collapse"))
		Expect(states.saved["cave"].Overrides.Lights).To(HaveKey("ambient"), "saved on exit")

		_, err = svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.Active()).To(Equal(scenes.Ref{SceneID: "cave", VariantID: "collapse"}), "restored")
		Expect(hue.applied[len(hue.applied)-1].instruction).To(Equal(scenes.LightInstruction{On: false}))
	})

	It("saves state on exit", func() {
		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).ToNot(HaveOccurred())
		saves := states.saves

		Expect(svc.Exit(ctx)).To(Succeed())
		Expect(states.saves).To(Equal(saves + 1))
		Expect(svc.Active().IsZero()).To(BeTrue())
	})

	It("requires an active scene for variant switches", func() {
		_, err := svc.SetVariant(ctx, "collapse")
		Expect(err).To(MatchError(application.ErrNoActiveScene))
	})

	It("reports devices without a driver", func() {
		room.AddDevice(&devices.Device{ID: "strip", Kind: devices.KindLight, Integration: devices.IntegrationMatter})
		Expect(room.Bind("ambient", "strip")).To(Succeed())

		_, err := svc.Activate(ctx, "cave", "")
		Expect(err).To(MatchError(application.ErrNoDriver))
		Expect(hue.applied).To(HaveLen(1), "other devices are still instructed")
	})
})
