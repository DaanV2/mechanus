package devices

// Device is a physical (or virtual) output in the room: a TV, a light, a speaker, an MQTT topic.
type Device struct {
	ID          string            `json:"id"`               // Unique identifier for the device
	Name        string            `json:"name"`             // Human readable name, e.g. "Table TV"
	Kind        Kind              `json:"kind"`             // What kind of instructions the device accepts
	Integration string            `json:"integration"`      // Driver used to reach the device, e.g. [IntegrationHue]
	Config      map[string]string `json:"config,omitempty"` // Integration specific settings (address, entity id, topic, ...)
}
