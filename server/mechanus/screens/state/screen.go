package state

import (
	"time"

	"github.com/DaanV2/mechanus/server/pkg/utility/must"
)

type ScreenState struct {
	Environment  *Environment  `json:"environment"`             // Environment state
	Grid         *GridSettings `json:"grid"`                    // Grid settings
	Layers       []*Layer      `json:"layers"`                  // List of layers, each containing entities
	SplashScreen *SplashScreen `json:"splash_screen,omitempty"` // Optional splash screen, if nil assumed to be not shown
}

func NewScreenState() *ScreenState {
	return &ScreenState{
		Environment: &Environment{
			CurrentTime:     must.Do(time.Parse("15:04", "12:00")),
			BackgroundColor: "#000000",
			LightIntensity:  1,
		},
		Grid: &GridSettings{
			XOffset: 0,
			YOffset: 0,
			XRatio:  1,
			YRatio:  1,
		},
		Layers: []*Layer{},
		SplashScreen: nil,
	}
}
