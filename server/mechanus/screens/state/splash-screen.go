package state

type SplashScreen struct {
	Show            bool   `json:"show"`             // Whether the splash screen is shown
	Title           string `json:"title"`            // Title of the splash screen
	Subtitle        string `json:"subtitle"`         // Subtitle of the splash screen
	BackgroundColor string `json:"background_color"` // Background color of the splash screen
}
