package application_test

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// bridgeDriver stands in for an integration that reaches its devices through a shared bridge or
// account, the way a light or music integration would: it translates its own instruction kind using
// the device config, and knows from the bridge whether a device is reachable.
type bridgeDriver struct {
	recordingDriver

	unreachable error
}

func (d *bridgeDriver) CheckDevice(context.Context, *devices.Device) error { return d.unreachable }

// flakyDriver has no way to check a device, and fails or blocks on command.
type flakyDriver struct {
	recordingDriver

	lock    sync.Mutex
	err     error
	blocked map[string]chan struct{}
}

func (d *flakyDriver) Apply(ctx context.Context, scene scenes.Ref, target string, device *devices.Device, ins scenes.Instruction) error {
	d.lock.Lock()
	err := d.err
	release := d.blocked[device.ID]
	d.lock.Unlock()

	if release != nil {
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err != nil {
		return err
	}

	return d.recordingDriver.Apply(ctx, scene, target, device, ins)
}

func (d *flakyDriver) setErr(err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.err = err
}

type recordingListener struct {
	transitions []*scenes.Transition
}

func (l *recordingListener) SceneChanged(_ context.Context, t *scenes.Transition) {
	l.transitions = append(l.transitions, t)
}

func statusOf(ctx context.Context, m *application.DeviceMonitor, id string) devices.Status {
	GinkgoHelper()

	all, err := m.Statuses(ctx)
	Expect(err).ToNot(HaveOccurred())
	for _, s := range all {
		if s.DeviceID == id {
			return s
		}
	}
	Fail("no status for device " + id)

	return devices.Status{}
}

var _ = Describe("DeviceMonitor", func() {
	var (
		ctx     context.Context
		room    *devices.Room
		configs configStore
		states  *stateStore
		lights  *bridgeDriver
		music   *flakyDriver
		monitor *application.DeviceMonitor
		svc     *application.SceneService
	)

	create := func(timeout time.Duration) {
		monitor = application.NewDeviceMonitor(staticRoom{room}, timeout)
		svc = application.NewSceneService(configs, states, staticRoom{room}, monitor.Track(lights), monitor.Track(music))
	}

	BeforeEach(func() {
		ctx = context.Background()
		room = devices.NewRoom("room", "Living room")
		room.AddDevice(&devices.Device{
			ID:          "lamp",
			Kind:        devices.KindLight,
			Integration: devices.IntegrationHue,
			Config:      map[string]string{"light_id": "3"},
		})
		room.AddDevice(&devices.Device{ID: "lamp-2", Kind: devices.KindLight, Integration: devices.IntegrationHue})
		room.AddDevice(&devices.Device{ID: "speaker", Kind: devices.KindMusic, Integration: devices.IntegrationSpotify})
		room.AddDevice(&devices.Device{ID: "fog", Kind: devices.KindCustom, Integration: devices.IntegrationMQTT})
		Expect(room.Bind("ambient", "lamp")).To(Succeed())
		Expect(room.Bind("ambient", "lamp-2")).To(Succeed())
		Expect(room.Bind("music", "speaker")).To(Succeed())

		configs = configStore{
			"tavern": {
				ID: "tavern",
				Instructions: scenes.Instructions{
					Lights: map[string]scenes.LightInstruction{"ambient": {On: true, Color: "#ffcc88", Brightness: 0.8}},
					Music:  map[string]scenes.MusicInstruction{"music": {Action: scenes.MusicPlay, Source: "spotify:playlist:tavern"}},
				},
				Variants: []scenes.Variant{{
					ID: "night",
					Instructions: scenes.Instructions{
						Lights: map[string]scenes.LightInstruction{"ambient": {On: true, Brightness: 0.1}},
					},
				}},
			},
		}
		states = &stateStore{saved: map[string]scenes.State{}}
		lights = &bridgeDriver{recordingDriver: recordingDriver{integration: devices.IntegrationHue}}
		music = &flakyDriver{recordingDriver: recordingDriver{integration: devices.IntegrationSpotify}, blocked: map[string]chan struct{}{}}
		create(0)
	})

	Context("with drivers for lights and music", func() {
		It("delivers each instruction kind with its target", func() {
			_, err := svc.Activate(ctx, "tavern", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(lights.applied).To(ConsistOf(
				applied{"lamp", "ambient", scenes.LightInstruction{On: true, Color: "#ffcc88", Brightness: 0.8}},
				applied{"lamp-2", "ambient", scenes.LightInstruction{On: true, Color: "#ffcc88", Brightness: 0.8}},
			))
			Expect(music.applied).To(ConsistOf(
				applied{"speaker", "music", scenes.MusicInstruction{Action: scenes.MusicPlay, Source: "spotify:playlist:tavern"}},
			))
		})

		It("reports a status for every device in the room, sorted", func() {
			all, err := monitor.Statuses(ctx)
			Expect(err).ToNot(HaveOccurred())

			ids := make([]string, 0, len(all))
			for _, s := range all {
				ids = append(ids, s.DeviceID)
			}
			Expect(ids).To(Equal([]string{"fog", "lamp", "lamp-2", "speaker"}))
		})
	})

	Describe("status", func() {
		It("is unknown until a device without a check has been instructed", func() {
			Expect(statusOf(ctx, monitor, "speaker").State).To(Equal(devices.StateUnknown))
		})

		It("follows the last instruction for a driver that cannot check", func() {
			_, err := svc.Activate(ctx, "tavern", "")
			Expect(err).ToNot(HaveOccurred())

			online := statusOf(ctx, monitor, "speaker")
			Expect(online.State).To(Equal(devices.StateOnline))
			Expect(online.LastApply).ToNot(BeZero())

			music.setErr(errors.New("speaker is gone"))
			_, err = svc.Adjust(ctx, scenes.Instructions{Music: map[string]scenes.MusicInstruction{"music": {Action: scenes.MusicPause}}})
			Expect(err).To(MatchError(ContainSubstring("speaker is gone")))

			offline := statusOf(ctx, monitor, "speaker")
			Expect(offline.State).To(Equal(devices.StateOffline))
			Expect(offline.Error).To(Equal("speaker is gone"))
			Expect(offline.LastApplyError).To(Equal("speaker is gone"))

			music.setErr(nil)
			_, err = svc.Adjust(ctx, scenes.Instructions{Music: map[string]scenes.MusicInstruction{"music": {Action: scenes.MusicStop}}})
			Expect(err).ToNot(HaveOccurred())

			recovered := statusOf(ctx, monitor, "speaker")
			Expect(recovered.State).To(Equal(devices.StateOnline))
			Expect(recovered.Error).To(BeEmpty())
			Expect(recovered.LastApplyError).To(BeEmpty())
		})

		It("takes the driver's own check over the last instruction", func() {
			Expect(statusOf(ctx, monitor, "lamp").State).To(Equal(devices.StateOnline), "checked, so known before any instruction")

			lights.unreachable = errors.New("bridge not reachable")
			status := statusOf(ctx, monitor, "lamp")
			Expect(status.State).To(Equal(devices.StateOffline))
			Expect(status.Error).To(Equal("bridge not reachable"))
		})

		It("says so when a device's integration has no driver", func() {
			status := statusOf(ctx, monitor, "fog")
			Expect(status.State).To(Equal(devices.StateOffline))
			Expect(status.Error).To(Equal(application.ErrNoDriver.Error()))
		})
	})

	Describe("a device that does not respond", func() {
		It("does not hold up the other devices", func() {
			release := make(chan struct{})
			music.blocked["speaker"] = release

			done := make(chan error, 1)
			go func() {
				_, err := svc.Activate(ctx, "tavern", "")
				done <- err
			}()

			Eventually(func(g Gomega) {
				lights.lock.Lock()
				defer lights.lock.Unlock()
				g.Expect(lights.applied).To(HaveLen(2))
			}).Should(Succeed())
			Consistently(done, "50ms").ShouldNot(Receive(), "still waiting on the speaker")

			close(release)
			Eventually(done).Should(Receive(Not(HaveOccurred())))
			Expect(music.applied).To(HaveLen(1))
		})

		It("counts as failed after the timeout", func() {
			create(time.Millisecond * 20)
			music.blocked["speaker"] = make(chan struct{})

			_, err := svc.Activate(ctx, "tavern", "")
			Expect(err).To(MatchError(context.DeadlineExceeded))
			Expect(err).To(MatchError(ContainSubstring(`device "speaker"`)))
			Expect(lights.applied).To(HaveLen(2))
			Expect(statusOf(ctx, monitor, "speaker").State).To(Equal(devices.StateOffline))
		})
	})
})

var _ = Describe("SceneService listeners", func() {
	var (
		ctx      context.Context
		svc      *application.SceneService
		listener *recordingListener
	)

	BeforeEach(func() {
		ctx = context.Background()
		configs := configStore{
			"tavern": {ID: "tavern", Variants: []scenes.Variant{{ID: "night"}}},
			"road":   {ID: "road"},
		}
		room := devices.NewRoom("room", "Living room")
		svc = application.NewSceneService(configs, &stateStore{saved: map[string]scenes.State{}}, staticRoom{room})
		listener = &recordingListener{}
		svc.AddListener(listener)
	})

	It("are told about scene and variant switches", func() {
		_, err := svc.Activate(ctx, "tavern", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.SetVariant(ctx, "night")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.Activate(ctx, "road", "")
		Expect(err).ToNot(HaveOccurred())

		Expect(listener.transitions).To(HaveLen(3))
		Expect(listener.transitions[0].From.IsZero()).To(BeTrue())
		Expect(listener.transitions[1].To).To(Equal(scenes.Ref{SceneID: "tavern", VariantID: "night"}))
		Expect(listener.transitions[2].From).To(Equal(scenes.Ref{SceneID: "tavern", VariantID: "night"}))
		Expect(listener.transitions[2].To).To(Equal(scenes.Ref{SceneID: "road"}))
	})

	It("are not told about changes within a variant", func() {
		_, err := svc.Activate(ctx, "tavern", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.Adjust(ctx, scenes.Instructions{Lights: map[string]scenes.LightInstruction{"ambient": {On: false}}})
		Expect(err).ToNot(HaveOccurred())
		_, err = svc.Activate(ctx, "tavern", "")
		Expect(err).ToNot(HaveOccurred())

		Expect(listener.transitions).To(HaveLen(1))
	})
})
