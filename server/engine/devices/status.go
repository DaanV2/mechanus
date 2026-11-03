package devices

import "time"

// State says whether a device can currently be reached.
type State string

// Device state constants.
const (
	StateUnknown State = "unknown" // StateUnknown means nothing is known yet: not instructed, and its driver cannot check it.
	StateOnline  State = "online"  // StateOnline means the device can be reached.
	StateOffline State = "offline" // StateOffline means the device cannot be reached, see [Status.Error].
)

// Status is the health / connection status of a device. It is runtime information, never stored.
type Status struct {
	DeviceID       string    `json:"device_id"`
	Integration    string    `json:"integration"`
	State          State     `json:"state"`
	Error          string    `json:"error,omitempty"`            // Why the device is offline
	LastApply      time.Time `json:"last_apply,omitzero"`        // When the device was last instructed, zero if never
	LastApplyError string    `json:"last_apply_error,omitempty"` // What went wrong the last time the device was instructed
}
