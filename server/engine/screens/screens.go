package screens

import (
	"context"
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
	screensv1 "github.com/DaanV2/mechanus/server/pkg/gen/proto/screens/v1"
)

type ListenerManager interface {
	Broadcast(msg *screensv1.ServerMessages) error
	Send(id string, msg *screensv1.ServerMessages) error
	HasListener(id string) bool
}

type Listener interface {
	Send(msg *screensv1.ServerMessages) error
}

type ScreenHandler struct {
	ID        string
	Listeners xsync.Map[string, Listener]
}

func (s *ScreenHandler) HandleMessages(ctx context.Context, listener Listener, msg *screensv1.ClientMessages) error {
	var response []*screensv1.ServerMessage
	var err error

	for _, v := range msg.GetAction() {
		resp, e := s.HandleMessage(ctx, v)
		if e != nil {
			err = errors.Join(err, e)
		}
		if len(resp) > 0 {
			response = append(response, resp...)
		}
	}

	if len(response) > 0 {
		var send screensv1.ServerMessages
		send.Action = response
		if e := listener.Send(&send); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func (s *ScreenHandler) HandleMessage(ctx context.Context, msg *screensv1.ClientMessage) ([]*screensv1.ServerMessage, error) {
	var response []*screensv1.ServerMessage
	var err error

	act := msg.GetAction()
	if act == nil {
		return response, nil
	}

	switch v := act.(type) {
	case *screensv1.ClientMessage_Ping:
		response, err = s.HandlePingRequest(ctx, v)
	case *screensv1.ClientMessage_InitialSetupRequest:
	}
	// Message track id
	id := msg.GetId()
	if id != "" {
		for _, resp := range response {
			resp.Id = &id
		}
	}

	return response, err
}

func (s *ScreenHandler) HandlePingRequest(ctx context.Context, msg *screensv1.ClientMessage_Ping) ([]*screensv1.ServerMessage, error) {
	return nil, nil // TODO implement handler later
}

func (s *ScreenHandler) HandleInitialSetupRequest(ctx context.Context, msg *screensv1.ClientMessage_InitialSetupRequest) ([]*screensv1.ServerMessage, error) {
	return nil, nil // TODO implement handler later
}
