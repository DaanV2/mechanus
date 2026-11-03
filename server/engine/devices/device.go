package devices

const maxIDLength = 128

// ValidID reports whether a client supplied device ID or target name is usable: it ends up in URLs and file names,
// so only short IDs of letters, digits, `-` and `_` are accepted.
func ValidID(id string) bool {
	if id == "" || len(id) > maxIDLength {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}

	return true
}

// Device is a physical (or virtual) output in the room: a TV, a light, a speaker, an MQTT topic.
type Device struct {
	ID          string            `json:"id"`               // Unique identifier for the device
	Name        string            `json:"name"`             // Human readable name, e.g. "Table TV"
	Kind        Kind              `json:"kind"`             // What kind of instructions the device accepts
	Integration string            `json:"integration"`      // Driver used to reach the device, e.g. [IntegrationHue]
	Config      map[string]string `json:"config,omitempty"` // Integration specific settings (address, entity id, topic, ...)
}
