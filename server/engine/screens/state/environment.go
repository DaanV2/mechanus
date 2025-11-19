package state

import "time"

// Environment represents the environmental conditions of a screen.
type Environment struct {
	CurrentTime     time.Time `json:"current_time"`     // Current time in the environment
	BackgroundColor string    `json:"background_color"` // Hex color code
	LightIntensity  float64   `json:"light_intensity"`  // Value between 0.0 (dark) and 1.0 (bright)
}
