package screens

import (
	"context"

	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
)

func (s *ScreenHandler) HandlePingRequest(ctx context.Context, msg *screensv1.ClientMessage_Ping) ([]*screensv1.ServerMessage, error) {
	s.Broadcast(ctx, &screensv1.ServerMessage{
		Action: &screensv1.ServerMessage_Ping{
			Ping: msg.Ping,
		},
	})

	return nil, nil // TODO implement handler later
}

func (s *ScreenHandler) HandleInitialSetupRequest(ctx context.Context, msg *screensv1.ClientMessage_InitialSetupRequest) ([]*screensv1.ServerMessage, error) {
	resp := []*screensv1.ServerMessage{
		&screensv1.ServerMessage{
			Action: screensv1.ServerMessage
		}
	}

	return resp, nil // TODO implement handler later
}
