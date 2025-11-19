package screens

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	"github.com/coder/websocket"
)

type ScreenHandler struct {
	id        string
	listeners *xsync.Map[string, *ScreenConn]
	logger    logging.Enriched
}

func NewScreenHandler(id string) *ScreenHandler {
	return &ScreenHandler{
		id:        id,
		listeners: xsync.NewMap[string, *ScreenConn](),
		logger:    logging.Enriched{}.WithPrefix("screen").With("screen_id", id),
	}
}

func (s *ScreenHandler) GetID() string {
	return s.id
}

// Sends a message to all listeners
func (s *ScreenHandler) Broadcast(ctx context.Context, msg ...*screensv1.ServerMessage) {
	msgs := &screensv1.ServerMessages{
		Action: msg,
	}

	for _, listener := range s.listeners.Items() {
		go s.broadcast(ctx, msgs, listener)
	}
}

func (s *ScreenHandler) broadcast(ctx context.Context, msg *screensv1.ServerMessages, listener *ScreenConn) {
	err := listener.Send(ctx, msg)
	if isClosed(err) {
		s.RemoveListener(listener.id)
	}
}

func (s *ScreenHandler) RemoveListener(id string) {
	s.listeners.Delete(id)
}

func (s *ScreenHandler) AddListener(ctx context.Context, listener *ScreenConn) {
	old, loaded := s.listeners.Swap(listener.id, listener)
	if loaded && old != nil {
		s.logger.From(ctx).Warn("A connection with this id already exists, closing old connection")
		go old.Close(errors.New("replaced by new connection"))
	}

	s.setupListener(listener)
}

func (s *ScreenHandler) setupListener(listener *ScreenConn) {
	defer listener.Close(nil)

	go listener.startReadLoop(s.HandleMessages)
}

func (s *ScreenHandler) HandleMessages(ctx context.Context, listener *ScreenConn, msg *screensv1.ClientMessages) error {
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
		if e := listener.Send(ctx, &send); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func (s *ScreenHandler) HandleMessage(ctx context.Context, msg *screensv1.ClientMessage) ([]*screensv1.ServerMessage, error) {
	// Non block read to see if we need to stop
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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
		response, err = s.HandleInitialSetupRequest(ctx, v)
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

func (s *ScreenHandler) Close() {
	wg := sync.WaitGroup{}

	for _, list := range s.listeners.Items() {
		wg.Go(func() {
			list.CloseWith(websocket.StatusGoingAway, "server closing")
		})
	}

	wg.Wait()
}

func isClosed(err error) bool {
	return errors.Is(err, net.ErrClosed)
}
