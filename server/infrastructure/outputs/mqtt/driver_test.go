package mqtt_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/infrastructure/outputs/mqtt"
	"github.com/charmbracelet/log"
	paho "github.com/eclipse/paho.mqtt.golang"
	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const driverClientID = "mechanus-test"

// reconnectWait covers the client's retry and backoff intervals.
const reconnectWait = time.Second * 15

// startBroker runs a real broker in the test process, returns it and the address it listens on.
func startBroker(address string) (*mochi.Server, string, error) {
	server := mochi.New(&mochi.Options{Logger: logging.ToSlog(log.New(io.Discard))})
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		return nil, "", err
	}

	listener := listeners.NewTCP(listeners.Config{ID: "test", Address: address})
	if err := server.AddListener(listener); err != nil {
		return nil, "", err
	}
	if err := server.Serve(); err != nil {
		return nil, "", err
	}

	return server, listener.Address(), nil
}

type received struct {
	topic    string
	payload  string
	retained bool
}

// subscriber is a device on the broker, recording all it receives.
type subscriber struct {
	client paho.Client
	lock   sync.Mutex
	got    []received
}

func subscribe(address, id string, topics ...string) *subscriber {
	GinkgoHelper()

	s := &subscriber{}
	opts := paho.NewClientOptions().AddBroker("tcp://" + address).SetClientID(id).SetOrderMatters(true)
	s.client = paho.NewClient(opts)

	token := s.client.Connect()
	Expect(token.WaitTimeout(time.Second * 5)).To(BeTrue())
	Expect(token.Error()).ToNot(HaveOccurred())

	for _, topic := range topics {
		sub := s.client.Subscribe(topic, 1, func(_ paho.Client, m paho.Message) {
			s.lock.Lock()
			defer s.lock.Unlock()
			s.got = append(s.got, received{m.Topic(), string(m.Payload()), m.Retained()})
		})
		Expect(sub.WaitTimeout(time.Second * 5)).To(BeTrue())
		Expect(sub.Error()).ToNot(HaveOccurred())
	}
	DeferCleanup(func() { s.client.Disconnect(0) })

	return s
}

func (s *subscriber) messages() []received {
	s.lock.Lock()
	defer s.lock.Unlock()

	return append([]received{}, s.got...)
}

var _ = Describe("Driver", func() {
	var (
		ctx        context.Context
		broker     *mochi.Server
		stopBroker func()
		address    string
		driver     *mqtt.Driver
		fog        *devices.Device
		ref        scenes.Ref
	)

	newDriver := func() *mqtt.Driver {
		GinkgoHelper()

		d, err := mqtt.NewDriver(&mqtt.Config{Broker: "tcp://" + address, ClientID: driverClientID})
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() { _ = d.AfterShutDown(context.Background()) })

		return d
	}

	connect := func(d *mqtt.Driver) {
		GinkgoHelper()

		Expect(d.AfterInitialize(ctx)).To(Succeed())
		Eventually(func() error { return d.CheckDevice(ctx, fog) }, "5s").Should(Succeed())
	}

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		broker, address, err = startBroker("127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		closeBroker := sync.OnceValue(broker.Close)
		stopBroker = func() { Expect(closeBroker()).To(Succeed()) }
		DeferCleanup(closeBroker)

		driver = newDriver()
		fog = &devices.Device{ID: "fog", Kind: devices.KindCustom, Integration: devices.IntegrationMQTT}
		ref = scenes.Ref{SceneID: "tavern", VariantID: "night"}
	})

	Describe("Topic", func() {
		It("defaults to the device's set topic", func() {
			Expect(mqtt.Topic(fog)).To(Equal("mechanus/devices/fog/set"))
		})

		It("comes from the device config when set", func() {
			fog.Config = map[string]string{mqtt.DeviceConfigTopic: "props/fog"}
			Expect(mqtt.Topic(fog)).To(Equal("props/fog"))
		})

		DescribeTable("rejects what cannot be published to",
			func(topic string) {
				fog.Config = map[string]string{mqtt.DeviceConfigTopic: topic}
				_, err := mqtt.Topic(fog)
				Expect(err).To(MatchError(mqtt.ErrInvalidTopic))
			},
			Entry("single level wildcard", "props/+/fog"),
			Entry("multi level wildcard", "props/#"),
		)
	})

	Describe("Apply", func() {
		BeforeEach(func() { connect(driver) })

		It("publishes the payload in an envelope", func() {
			device := subscribe(address, "device", "mechanus/devices/fog/set")

			instruction := scenes.CustomInstruction{Payload: json.RawMessage(`{"burst":3,"keep":"as is"}`)}
			Expect(driver.Apply(ctx, ref, "smoke", fog, instruction)).To(Succeed())

			Eventually(device.messages).Should(HaveLen(1))
			Expect(device.messages()[0].payload).To(MatchJSON(`{
				"scene": "tavern",
				"variant": "night",
				"target": "smoke",
				"payload": {"burst": 3, "keep": "as is"}
			}`))
		})

		It("publishes retained, so a device that connects later gets it", func() {
			Expect(driver.Apply(ctx, ref, "smoke", fog, scenes.CustomInstruction{Payload: json.RawMessage(`1`)})).To(Succeed())
			Expect(driver.Apply(ctx, ref, "smoke", fog, scenes.CustomInstruction{Payload: json.RawMessage(`2`)})).To(Succeed())

			late := subscribe(address, "late", "mechanus/devices/fog/set")

			Eventually(late.messages).Should(HaveLen(1), "only the latest is retained")
			Expect(late.messages()[0].retained).To(BeTrue())
			Expect(late.messages()[0].payload).To(ContainSubstring(`"payload":2`))
		})

		It("publishes to the topic from the device config", func() {
			fog.Config = map[string]string{mqtt.DeviceConfigTopic: "props/fog"}
			device := subscribe(address, "device", "props/fog")

			Expect(driver.Apply(ctx, ref, "smoke", fog, scenes.CustomInstruction{Payload: json.RawMessage(`1`)})).To(Succeed())

			Eventually(device.messages).Should(HaveLen(1))
		})

		It("ignores other kinds of instructions", func() {
			device := subscribe(address, "device", "#")
			Eventually(device.messages).Should(HaveLen(1), "the retained status")

			Expect(driver.Apply(ctx, ref, "smoke", fog, scenes.LightInstruction{On: true})).To(Succeed())

			Consistently(device.messages, "100ms").Should(HaveLen(1))
		})
	})

	Describe("with one failing device in a scene", func() {
		It("still instructs the other devices", func() {
			connect(driver)

			room := devices.NewRoom("room", "Room")
			room.AddDevice(&devices.Device{ID: "a-broken", Kind: devices.KindCustom, Integration: devices.IntegrationMQTT, Config: map[string]string{mqtt.DeviceConfigTopic: "props/#"}})
			room.AddDevice(fog)
			Expect(room.Bind("smoke", "a-broken")).To(Succeed())
			Expect(room.Bind("smoke", "fog")).To(Succeed())

			monitor := application.NewDeviceMonitor(staticRoom{room}, time.Second)
			service := application.NewSceneService(
				staticScene{&scenes.Config{ID: "tavern", Instructions: scenes.Instructions{
					Custom: map[string]scenes.CustomInstruction{"smoke": {Payload: json.RawMessage(`{"burst":1}`)}},
				}}},
				noState{}, staticRoom{room}, monitor.Track(driver),
			)

			device := subscribe(address, "device", "mechanus/devices/fog/set")
			transition, err := service.Activate(ctx, "tavern", "")

			Expect(err).To(MatchError(mqtt.ErrInvalidTopic))
			Expect(err).To(MatchError(ContainSubstring(`device "a-broken"`)))
			Expect(transition).ToNot(BeNil(), "the scene is active regardless")
			Eventually(device.messages).Should(HaveLen(1))

			statuses, err := monitor.Statuses(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses).To(HaveLen(2))
			Expect(statuses[0].State).To(Equal(devices.StateOffline))
			Expect(statuses[1].State).To(Equal(devices.StateOnline))
		})
	})

	Describe("SceneChanged", func() {
		BeforeEach(func() { connect(driver) })

		It("publishes the change to the scene event topic, not retained", func() {
			listener := subscribe(address, "listener", mqtt.TopicSceneEvents)

			driver.SceneChanged(ctx, &scenes.Transition{From: scenes.Ref{SceneID: "road"}, To: ref})

			Eventually(listener.messages).Should(HaveLen(1))
			Expect(listener.messages()[0].payload).To(MatchJSON(`{
				"scene": "tavern",
				"variant": "night",
				"from": {"scene": "road", "variant": ""}
			}`))

			late := subscribe(address, "late", mqtt.TopicSceneEvents)
			Consistently(late.messages, "100ms").Should(BeEmpty())
		})

		It("leaves out where it came from when no scene was active", func() {
			listener := subscribe(address, "listener", mqtt.TopicSceneEvents)

			driver.SceneChanged(ctx, &scenes.Transition{To: ref})

			Eventually(listener.messages).Should(HaveLen(1))
			Expect(listener.messages()[0].payload).To(MatchJSON(`{"scene": "tavern", "variant": "night"}`))
		})
	})

	Describe("availability", func() {
		It("announces online, retained", func() {
			connect(driver)

			late := subscribe(address, "late", mqtt.TopicStatus)
			Eventually(late.messages).Should(ContainElement(received{mqtt.TopicStatus, mqtt.StatusOnline, true}))
		})

		It("has the broker publish its last will when the connection drops", func() {
			connect(driver)
			status := subscribe(address, "status", mqtt.TopicStatus)
			Eventually(status.messages).Should(ConsistOf(received{mqtt.TopicStatus, mqtt.StatusOnline, true}))

			client, ok := broker.Clients.Get(driverClientID)
			Expect(ok).To(BeTrue())
			client.Stop(errors.New("connection dropped by test"))

			Eventually(status.messages, "5s").Should(ContainElement(received{mqtt.TopicStatus, mqtt.StatusOffline, false}))
		})

		It("announces offline itself on a clean shutdown", func() {
			connect(driver)
			status := subscribe(address, "status", mqtt.TopicStatus)
			Eventually(status.messages).Should(HaveLen(1))

			Expect(driver.AfterShutDown(ctx)).To(Succeed())

			Eventually(status.messages).Should(HaveLen(2))
			Expect(status.messages()[1].payload).To(Equal(mqtt.StatusOffline))

			late := subscribe(address, "late", mqtt.TopicStatus)
			Eventually(late.messages).Should(ConsistOf(received{mqtt.TopicStatus, mqtt.StatusOffline, true}))
		})
	})

	Describe("when the broker is down", func() {
		It("does not hold up starting the server", func() {
			stopBroker()
			Expect(driver.AfterInitialize(ctx)).To(Succeed())
			Eventually(func() error { return driver.CheckDevice(ctx, fog) }, "5s").Should(MatchError(ContainSubstring("connect")))
		})

		It("fails right away", func() {
			start := time.Now()
			err := driver.Apply(ctx, ref, "smoke", fog, scenes.CustomInstruction{Payload: json.RawMessage(`1`)})

			Expect(err).To(MatchError(mqtt.ErrNotConnected))
			Expect(time.Since(start)).To(BeNumerically("<", time.Second))
			Expect(driver.CheckDevice(ctx, fog)).To(MatchError(mqtt.ErrNotConnected))
		})

		It("delivers the latest instruction per topic once it is back", func() {
			connect(driver)
			stopBroker()
			Eventually(func() error { return driver.CheckDevice(ctx, fog) }, "5s").Should(MatchError(mqtt.ErrNotConnected))

			instruct := func(payload string) {
				GinkgoHelper()
				err := driver.Apply(ctx, ref, "smoke", fog, scenes.CustomInstruction{Payload: json.RawMessage(payload)})
				Expect(err).To(MatchError(mqtt.ErrNotConnected))
			}
			instruct(`"old"`)
			instruct(`"new"`)

			var restarted *mochi.Server
			Eventually(func() (err error) {
				restarted, _, err = startBroker(address)

				return err
			}, "5s", "100ms").Should(Succeed())
			DeferCleanup(func() { _ = restarted.Close() })

			Eventually(func() error { return driver.CheckDevice(ctx, fog) }, reconnectWait).Should(Succeed())

			device := subscribe(address, "device", "mechanus/devices/fog/set")
			Eventually(device.messages).Should(HaveLen(1))
			Expect(device.messages()[0].retained).To(BeTrue())
			Expect(device.messages()[0].payload).To(ContainSubstring(`"payload":"new"`))
		})
	})

	Describe("Config", func() {
		DescribeTable("Validate",
			func(broker string, valid bool) {
				cfg := mqtt.Config{Broker: broker}
				if valid {
					Expect(cfg.Validate()).To(Succeed())
				} else {
					Expect(cfg.Validate()).To(MatchError(mqtt.ErrInvalidBroker))
				}
			},
			Entry("tcp", "tcp://192.168.1.10:1883", true),
			Entry("tls", "tls://broker.local:8883", true),
			Entry("websocket", "ws://broker.local:9001", true),
			Entry("no scheme", "192.168.1.10:1883", false),
			Entry("no host", "tcp://", false),
			Entry("unknown scheme", "http://broker.local", false),
		)

		It("leaves a password in the address out of the error", func() {
			const secret = "hunter2"
			cfg := mqtt.Config{Broker: "http://user:" + secret + "@broker.local"}
			Expect(cfg.Validate().Error()).ToNot(ContainSubstring(secret))
		})

		It("is disabled without a broker", func() {
			cfg := mqtt.Config{}
			Expect(cfg.Enabled()).To(BeFalse())
		})
	})
})

type staticRoom struct{ room *devices.Room }

func (r staticRoom) CurrentRoom(context.Context) (*devices.Room, error) { return r.room, nil }

// staticScene is a store holding a single scene.
type staticScene struct{ cfg *scenes.Config }

func (s staticScene) GetSceneConfig(context.Context, string) (*scenes.Config, error) {
	return s.cfg, nil
}

func (s staticScene) ListSceneConfigs(ctx context.Context) ([]*scenes.Config, error) {
	cfg, err := s.GetSceneConfig(ctx, s.cfg.ID)

	return []*scenes.Config{cfg}, err
}

// noState is a store for scenes that are never saved.
type noState struct{}

func (noState) LoadSceneState(context.Context, string) (*scenes.State, error) {
	return nil, scenes.ErrStateNotFound
}

func (noState) SaveSceneState(context.Context, *scenes.State) error { return nil }
