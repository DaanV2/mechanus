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

type sceneRepo map[string]*scenes.Scene

func (r sceneRepo) GetScene(_ context.Context, id string) (*scenes.Scene, error) {
	if s, ok := r[id]; ok {
		return s, nil
	}

	return nil, errors.New("not found")
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
		room   *devices.Room
		screen *recordingDriver
		hue    *recordingDriver
		svc    *application.SceneService
	)

	BeforeEach(func() {
		room = devices.NewRoom("room", "Living room")
		room.AddDevice(&devices.Device{ID: "tv", Kind: devices.KindScreen, Integration: devices.IntegrationScreen})
		room.AddDevice(&devices.Device{ID: "lamp", Kind: devices.KindLight, Integration: devices.IntegrationHue})
		Expect(room.Bind("table", "tv")).To(Succeed())
		Expect(room.Bind("ambient", "lamp")).To(Succeed())

		repo := sceneRepo{"cave": {
			ID: "cave",
			Instructions: scenes.Instructions{
				Screens: map[string]scenes.ScreenInstruction{"table": {View: scenes.ViewBattleMap, Asset: "cave.dd2vtt"}},
				Lights:  map[string]scenes.LightInstruction{"ambient": {On: true, Brightness: 0.2}},
			},
			Variants: []scenes.Variant{{
				ID: "collapse",
				Instructions: scenes.Instructions{
					Lights: map[string]scenes.LightInstruction{"ambient": {On: true, Brightness: 0.2, Effect: "flicker"}},
				},
			}},
		}}

		screen = &recordingDriver{integration: devices.IntegrationScreen}
		hue = &recordingDriver{integration: devices.IntegrationHue}
		svc = application.NewSceneService(repo, staticRoom{room}, screen, hue)
	})

	It("instructs bound devices through their driver", func() {
		_, err := svc.Activate(context.Background(), "cave", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.Current()).To(Equal(scenes.State{SceneID: "cave"}))
		Expect(screen.applied).To(HaveLen(1))
		Expect(screen.applied[0].device).To(Equal("tv"))
		Expect(hue.applied).To(HaveLen(1))
	})

	It("only re-instructs changed devices on a variant transition", func() {
		_, err := svc.Activate(context.Background(), "cave", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.Activate(context.Background(), "cave", "collapse")
		Expect(err).ToNot(HaveOccurred())

		Expect(screen.applied).To(HaveLen(1))
		Expect(hue.applied).To(HaveLen(2))
		Expect(hue.applied[1].instruction).To(Equal(scenes.LightInstruction{On: true, Brightness: 0.2, Effect: "flicker"}))
	})

	It("reports devices without a driver", func() {
		room.AddDevice(&devices.Device{ID: "strip", Kind: devices.KindLight, Integration: devices.IntegrationMatter})
		Expect(room.Bind("ambient", "strip")).To(Succeed())

		_, err := svc.Activate(context.Background(), "cave", "")
		Expect(err).To(MatchError(application.ErrNoDriver))
		Expect(hue.applied).To(HaveLen(1), "other devices are still instructed")
	})
})
