package devices

// Kind is the category of output a device provides. Scene instructions are typed per kind, so a
// device only ever receives instructions it can act on.
type Kind string

// Device kind constants.
const (
	KindScreen Kind = "screen" // KindScreen renders a view (battle map, scenery, info board, custom).
	KindLight  Kind = "light"  // KindLight is a light or group of lights.
	KindMusic  Kind = "music"  // KindMusic is a music / audio player.
	KindCustom Kind = "custom" // KindCustom is anything else, fed raw payloads over an event bus.
)

// Well-known integration identifiers. An integration names the driver that knows how to talk to a
// device. These are identifiers only; the drivers themselves live in infrastructure and most are
// not implemented yet.
const (
	IntegrationScreen        = "screen"        // Built-in browser screen client (WebSocket).
	IntegrationHue           = "hue"           // Philips Hue bridge.
	IntegrationMatter        = "matter"        // Matter devices.
	IntegrationHomeAssistant = "homeassistant" // Home Assistant entities.
	IntegrationSpotify       = "spotify"       // Spotify Connect.
	IntegrationITunes        = "itunes"        // iTunes / Apple Music.
	IntegrationMQTT          = "mqtt"          // MQTT broker, for custom devices.
)
