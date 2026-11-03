// Package mqtt implements the output driver for custom devices behind an MQTT broker. A scene holds
// an opaque payload per target; the driver wraps it in an [Envelope] and publishes it, retained, to
// the topic of each device bound to that target. It also announces scene changes and its own
// availability, see [TopicSceneEvents] and [TopicStatus].
package mqtt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	// TopicStatus receives the availability of the server, retained: [StatusOnline] or [StatusOffline].
	// The broker publishes [StatusOffline] itself (last will) when the server disappears.
	TopicStatus = "mechanus/status"
	// TopicSceneEvents receives a [SceneEvent] for every scene or variant change.
	TopicSceneEvents = "mechanus/events/scene"
	// StatusOnline is published to [TopicStatus] when the server connects.
	StatusOnline = "online"
	// StatusOffline is published to [TopicStatus] when the server goes away.
	StatusOffline = "offline"

	// DeviceConfigTopic is the [devices.Device.Config] key that overrides the topic of a device.
	DeviceConfigTopic = "topic"

	qosAtLeastOnce = byte(1)
	// publishTimeout limits how long a publish waits for the broker's acknowledgement.
	publishTimeout = time.Second * 5
	// connectTimeout limits how long a connection attempt waits for the broker to answer. Retries only
	// happen after it, so a broker that accepts a connection but never answers holds up reconnecting.
	connectTimeout = time.Second * 5
	// disconnectQuiesce is how long, in milliseconds, in flight messages get when disconnecting.
	disconnectQuiesce = 250
)

var (
	// ErrNotConnected is returned when the broker cannot be reached.
	ErrNotConnected = errors.New("not connected to the mqtt broker")
	// ErrUndelivered is reported for a device whose last instruction has not reached the broker. It is
	// delivered once the connection is back.
	ErrUndelivered = errors.New("last instruction not delivered to the mqtt broker")
	// ErrInvalidTopic is returned when a device's topic cannot be published to.
	ErrInvalidTopic = errors.New("invalid mqtt topic")
)

var (
	_ application.OutputDriver  = &Driver{}
	_ application.DeviceChecker = &Driver{}
	_ application.SceneListener = &Driver{}
)

// Envelope is what a device receives on its topic.
type Envelope struct {
	Scene   string          `json:"scene"`
	Variant string          `json:"variant"` // Empty means the scene's base instructions
	Target  string          `json:"target"`
	Payload json.RawMessage `json:"payload"` // The scene's payload for the target, untouched
}

// SceneEvent is published to [TopicSceneEvents] when the active scene or variant changes.
type SceneEvent struct {
	Scene   string     `json:"scene"`
	Variant string     `json:"variant"`
	From    *SceneFrom `json:"from,omitempty"` // Not set when no scene was active
}

// SceneFrom is the scene and variant a [SceneEvent] moved away from.
type SceneFrom struct {
	Scene   string `json:"scene"`
	Variant string `json:"variant"`
}

// Driver delivers custom instructions to devices over an MQTT broker. A broker that is down does not
// stop the server: the driver keeps reconnecting, instructions fail right away, and the latest
// undelivered instruction per topic is published once the connection is back. Retained messages make
// that the state a device finds when it (re)connects.
type Driver struct {
	client paho.Client

	lock    sync.Mutex
	pending map[string][]byte // Topic -> latest instruction that has not reached the broker
	lastErr error             // Why the last connection attempt failed
}

// NewDriver creates the driver. It does not connect until [Driver.AfterInitialize].
func NewDriver(cfg *Config) (*Driver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	d := &Driver{pending: map[string][]byte{}}

	opts := paho.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetWill(TopicStatus, StatusOffline, qosAtLeastOnce, true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(time.Second * 5).
		SetConnectTimeout(connectTimeout).
		SetMaxReconnectInterval(time.Second * 30).
		SetOnConnectHandler(d.onConnect).
		SetConnectionNotificationHandler(d.onConnectionNotification)
	d.client = paho.NewClient(opts)

	return d, nil
}

// AfterInitialize implements lifecycle.AfterInitialize: starts connecting in the background, a broker
// that is not up yet should not hold up the server.
func (d *Driver) AfterInitialize(ctx context.Context) error {
	logging.From(ctx).Debug("connecting to the mqtt broker")
	d.client.Connect()

	return nil
}

// AfterShutDown implements lifecycle.AfterShutDown: a clean disconnect does not trigger the last will,
// so the driver announces it is going offline itself.
func (d *Driver) AfterShutDown(ctx context.Context) error {
	var err error
	if d.client.IsConnectionOpen() {
		err = d.publish(ctx, TopicStatus, true, []byte(StatusOffline))
	}
	d.client.Disconnect(disconnectQuiesce)

	return err
}

// Integration implements application.OutputDriver.
func (d *Driver) Integration() string {
	return devices.IntegrationMQTT
}

// Apply implements application.OutputDriver. Instructions of other kinds are ignored.
func (d *Driver) Apply(ctx context.Context, scene scenes.Ref, target string, device *devices.Device, instruction scenes.Instruction) error {
	ins, ok := instruction.(scenes.CustomInstruction)
	if !ok {
		return nil
	}

	topic, err := Topic(device)
	if err != nil {
		return err
	}
	body, err := json.Marshal(Envelope{Scene: scene.SceneID, Variant: scene.VariantID, Target: target, Payload: ins.Payload})
	if err != nil {
		return fmt.Errorf("creating envelope: %w", err)
	}

	token := d.publishInstruction(topic, body)
	if token == nil {
		return ErrNotConnected
	}
	if err := wait(ctx, token); err != nil {
		return fmt.Errorf("publishing to %q: %w", topic, err)
	}
	d.delivered(topic, body)

	return nil
}

// CheckDevice implements application.DeviceChecker.
func (d *Driver) CheckDevice(_ context.Context, device *devices.Device) error {
	topic, err := Topic(device)
	if err != nil {
		return err
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.client.IsConnectionOpen() {
		if d.lastErr != nil {
			return fmt.Errorf("%w: %w", ErrNotConnected, d.lastErr)
		}

		return ErrNotConnected
	}
	if _, ok := d.pending[topic]; ok {
		return ErrUndelivered
	}

	return nil
}

// SceneChanged implements application.SceneListener. Events are about the moment they happen, so they
// are not retained and not kept for later when the broker is down.
func (d *Driver) SceneChanged(ctx context.Context, transition *scenes.Transition) {
	event := SceneEvent{Scene: transition.To.SceneID, Variant: transition.To.VariantID}
	if !transition.From.IsZero() {
		event.From = &SceneFrom{Scene: transition.From.SceneID, Variant: transition.From.VariantID}
	}

	body, err := json.Marshal(event)
	if err == nil {
		err = d.publish(ctx, TopicSceneEvents, false, body)
	}
	if err != nil {
		logging.From(ctx).Warn("could not publish scene event to mqtt", "error", err)
	}
}

// Topic returns the topic instructions for a device are published to.
func Topic(device *devices.Device) (string, error) {
	topic := device.Config[DeviceConfigTopic]
	if topic == "" {
		topic = "mechanus/devices/" + device.ID + "/set"
	}
	if strings.ContainsAny(topic, "+#\x00") {
		return "", fmt.Errorf("%w %q: wildcards cannot be published to", ErrInvalidTopic, topic)
	}

	return topic, nil
}

// publishInstruction remembers an instruction until it is delivered and hands it to the client, or
// returns nil when there is no connection. The lock keeps the order the client sends messages in the
// same as the order of pending, so a redelivery cannot overtake a newer instruction.
func (d *Driver) publishInstruction(topic string, body []byte) paho.Token {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.pending[topic] = body
	if !d.client.IsConnectionOpen() {
		return nil
	}

	return d.client.Publish(topic, qosAtLeastOnce, true, body)
}

// delivered forgets an instruction, unless a newer one for the same topic took its place.
func (d *Driver) delivered(topic string, body []byte) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if bytes.Equal(d.pending[topic], body) {
		delete(d.pending, topic)
	}
}

func (d *Driver) publish(ctx context.Context, topic string, retained bool, body []byte) error {
	if !d.client.IsConnectionOpen() {
		return ErrNotConnected
	}

	return wait(ctx, d.client.Publish(topic, qosAtLeastOnce, retained, body))
}

// onConnect runs on every (re)connect, in its own goroutine.
func (d *Driver) onConnect(_ paho.Client) {
	ctx := context.Background()
	logger := logging.From(ctx)
	logger.Info("connected to the mqtt broker")

	d.lock.Lock()
	d.lastErr = nil
	d.lock.Unlock()

	if err := d.publish(ctx, TopicStatus, true, []byte(StatusOnline)); err != nil {
		logger.Warn("could not publish mqtt status", "error", err)
	}
	for topic, err := range d.redeliver(ctx) {
		logger.Warn("could not redeliver mqtt instruction", "topic", topic, "error", err)
	}
}

// redeliver publishes the instructions that did not reach the broker and returns what still failed.
func (d *Driver) redeliver(ctx context.Context) map[string]error {
	type message struct {
		body  []byte
		token paho.Token
	}

	d.lock.Lock()
	messages := make(map[string]message, len(d.pending))
	for topic, body := range d.pending {
		messages[topic] = message{body, d.client.Publish(topic, qosAtLeastOnce, true, body)}
	}
	d.lock.Unlock()

	failed := map[string]error{}
	for topic, m := range messages {
		if err := wait(ctx, m.token); err != nil {
			failed[topic] = err

			continue
		}
		d.delivered(topic, m.body)
	}

	return failed
}

func (d *Driver) onConnectionNotification(_ paho.Client, notification paho.ConnectionNotification) {
	logger := logging.From(context.Background())

	switch n := notification.(type) {
	case paho.ConnectionNotificationLost:
		logger.Warn("lost the connection to the mqtt broker", "error", n.Reason)
	case paho.ConnectionNotificationBrokerFailed:
		d.lock.Lock()
		changed := d.lastErr == nil || d.lastErr.Error() != n.Reason.Error()
		d.lastErr = n.Reason
		d.lock.Unlock()

		// Retried for as long as the broker is down, so only tell when the reason changes.
		if changed {
			logger.Warn("could not connect to the mqtt broker, retrying", "error", n.Reason)
		}
	}
}

// wait blocks until the broker acknowledged a message, publishTimeout passed, or ctx is done.
func wait(ctx context.Context, token paho.Token) error {
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}
