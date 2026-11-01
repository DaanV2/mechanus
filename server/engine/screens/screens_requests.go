package screens

import (
	"context"

	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
)

// HandlePingRequest processes a ping request from a client.
func (s *ScreenHandler) HandlePingRequest(ctx context.Context, msg *screensv1.ClientMessage_Ping) ([]*screensv1.ServerMessage, error) {
	s.Broadcast(ctx, &screensv1.ServerMessage{
		Action: &screensv1.ServerMessage_Ping{
			Ping: msg.Ping,
		},
	})

	return nil, nil // TODO implement handler later
}

// HandleInitialSetupRequest processes an initial setup request from a client.
func (s *ScreenHandler) HandleInitialSetupRequest(ctx context.Context, msg *screensv1.ClientMessage_InitialSetupRequest) ([]*screensv1.ServerMessage, error) {
	st := s.State()
	response := make([]*screensv1.ServerMessage, 0, 2)
	response = append(response, &screensv1.ServerMessage{
		Action: &screensv1.ServerMessage_InitialSetup{
			InitialSetup: &screensv1.InitialSetupResponse{
				State: st,
			},
		},
	})
	// The splash screen is only shown while no scene has instructed this screen yet
	if st.GetView().GetKind() != screensv1.ViewKind_VIEW_KIND_BLANK {
		return response, nil
	}

	return append(response,
		&screensv1.ServerMessage{
			// TODO add splash screen data to the splashscreen update
			Action: &screensv1.ServerMessage_SplashScreenUpdate{
				SplashScreenUpdate: &screensv1.SplashScreen{
					Show:          true,
					Title:         new("MECHANUS"),
					Subtitle:      new(""),
					BackgroundHex: new("#000000"),
				},
			},
		},
	), nil
}
