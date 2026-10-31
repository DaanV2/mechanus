package devices

import (
	"errors"
	"slices"
)

var (
	// ErrDeviceNotFound is returned when a device is not part of the room.
	ErrDeviceNotFound = errors.New("device not found")
)

// Room is the physical setup scenes are played in. Scenes never reference devices directly; they
// address logical targets (e.g. "table", "wall", "ambient", "music"). The room binds those targets
// to the devices that happen to be present, so the same scene works in a different setup.
// A room is configuration, loaded from file storage.
type Room struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Devices  map[string]*Device  `json:"devices"`  // Device ID -> device
	Bindings map[string][]string `json:"bindings"` // Target -> device IDs
}

// NewRoom creates an empty room.
func NewRoom(id, name string) *Room {
	return &Room{
		ID:       id,
		Name:     name,
		Devices:  map[string]*Device{},
		Bindings: map[string][]string{},
	}
}

// GetID returns the room ID.
func (r *Room) GetID() string {
	return r.ID
}

// AddDevice adds or replaces a device in the room.
func (r *Room) AddDevice(device *Device) {
	r.Devices[device.ID] = device
}

// Bind assigns a device to a target. A target can have multiple devices and a device can serve
// multiple targets.
func (r *Room) Bind(target, deviceID string) error {
	if _, ok := r.Devices[deviceID]; !ok {
		return ErrDeviceNotFound
	}
	if !slices.Contains(r.Bindings[target], deviceID) {
		r.Bindings[target] = append(r.Bindings[target], deviceID)
	}

	return nil
}

// DevicesFor returns the devices of the given kind bound to a target.
func (r *Room) DevicesFor(target string, kind Kind) []*Device {
	var result []*Device
	for _, id := range r.Bindings[target] {
		if d, ok := r.Devices[id]; ok && d.Kind == kind {
			result = append(result, d)
		}
	}

	return result
}

// TargetsFor returns the targets a device is bound to.
func (r *Room) TargetsFor(deviceID string) []string {
	var result []string
	for target, ids := range r.Bindings {
		if slices.Contains(ids, deviceID) {
			result = append(result, target)
		}
	}
	slices.Sort(result)

	return result
}
