package screens

type Listener interface {
}

type ScreenListeners struct {
	// map the screen ID to its listeners
	listeners map[ScreenID][]Listener
}
