package screens

type ScreenListeners struct {
	// map the screen ID to its listeners
	listeners map[string]struct{}
}