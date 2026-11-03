package application

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
)

// DeviceChecker is optionally implemented by an [OutputDriver] that can tell whether a device is
// reachable: the broker or bridge connection is up, the light reports reachable, the speaker is
// available, ... The answer is authoritative for the device's status. Without it the status follows
// the outcome of the last instruction.
type DeviceChecker interface {
	// CheckDevice returns why the device cannot be reached, nil when it can. It is called when the status
	// is read, so it should be cheap, and it must honour ctx.
	CheckDevice(ctx context.Context, device *devices.Device) error
}

type applyResult struct {
	at  time.Time
	err error
}

// DeviceMonitor tracks the health / connection status of every device in the room. Drivers are
// registered through [DeviceMonitor.Track], which also limits how long a single instruction can take.
type DeviceMonitor struct {
	rooms   RoomConfigStore
	timeout time.Duration

	lock    sync.Mutex
	drivers map[string]OutputDriver
	results map[string]applyResult
}

// NewDeviceMonitor creates a monitor for the devices in the room. Every instruction gets at most
// timeout to complete, 0 means no limit.
func NewDeviceMonitor(rooms RoomConfigStore, timeout time.Duration) *DeviceMonitor {
	return &DeviceMonitor{
		rooms:   rooms,
		timeout: timeout,
		drivers: map[string]OutputDriver{},
		results: map[string]applyResult{},
	}
}

// Track registers a driver and returns it wrapped, pass the result to [NewSceneService].
func (m *DeviceMonitor) Track(driver OutputDriver) OutputDriver {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.drivers[driver.Integration()] = driver

	return &trackedDriver{OutputDriver: driver, monitor: m}
}

// Statuses returns the status of every device in the room, sorted by device ID. A device whose
// integration has no driver (not configured) is offline with [ErrNoDriver] as the reason.
func (m *DeviceMonitor) Statuses(ctx context.Context) ([]devices.Status, error) {
	room, err := m.rooms.CurrentRoom(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading room: %w", err)
	}

	result := make([]devices.Status, 0, len(room.Devices))
	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, device := range room.Devices {
		wg.Go(func() {
			status := m.status(ctx, device)

			lock.Lock()
			defer lock.Unlock()
			result = append(result, status)
		})
	}
	wg.Wait()

	slices.SortFunc(result, func(a, b devices.Status) int { return cmp.Compare(a.DeviceID, b.DeviceID) })

	return result, nil
}

func (m *DeviceMonitor) status(ctx context.Context, device *devices.Device) devices.Status {
	m.lock.Lock()
	driver, hasDriver := m.drivers[device.Integration]
	last, applied := m.results[device.ID]
	m.lock.Unlock()

	status := devices.Status{DeviceID: device.ID, Integration: device.Integration, State: devices.StateUnknown}
	if applied {
		status.LastApply = last.at
		status.State = devices.StateOnline
		if last.err != nil {
			status.LastApplyError = last.err.Error()
			status.State = devices.StateOffline
			status.Error = last.err.Error()
		}
	}

	if !hasDriver {
		status.State = devices.StateOffline
		status.Error = ErrNoDriver.Error()

		return status
	}

	checker, ok := driver.(DeviceChecker)
	if !ok {
		return status
	}

	ctx, cancel := m.withTimeout(ctx)
	defer cancel()

	status.State = devices.StateOnline
	status.Error = ""
	if err := checker.CheckDevice(ctx, device); err != nil {
		status.State = devices.StateOffline
		status.Error = err.Error()
	}

	return status
}

func (m *DeviceMonitor) record(deviceID string, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.results[deviceID] = applyResult{at: time.Now(), err: err}
}

func (m *DeviceMonitor) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if m.timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, m.timeout)
}

// trackedDriver limits how long an instruction can take and records its outcome.
type trackedDriver struct {
	OutputDriver

	monitor *DeviceMonitor
}

// Apply implements [OutputDriver].
func (t *trackedDriver) Apply(ctx context.Context, scene scenes.Ref, target string, device *devices.Device, instruction scenes.Instruction) error {
	ctx, cancel := t.monitor.withTimeout(ctx)
	defer cancel()

	err := t.OutputDriver.Apply(ctx, scene, target, device, instruction)
	t.monitor.record(device.ID, err)

	return err
}
