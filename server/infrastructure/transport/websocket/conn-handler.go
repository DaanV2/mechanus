package websocket

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/DaanV2/mechanus/server/mechanus/screens"
	screensv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1"
	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

type connectionHandler struct {
	id       string // user / device id
	screenid string // screen id

	ctx     context.Context
	conn    *websocket.Conn
	closeFn func()
	logger  *log.Logger
	info    *ConnectionInfo
	handler *screens.ScreenHandler
}

// Context implements WebsocketInfo.
func (connh *connectionHandler) Context() context.Context {
	return connh.ctx
}

// GetInfo implements WebsocketInfo.
func (connh *connectionHandler) GetInfo() *ConnectionInfo {
	return connh.info
}

// Write implements WebsocketInfo.
func (connh *connectionHandler) Write(p []byte) error {
	return connh.conn.Write(connh.ctx, websocket.MessageBinary, p)
}

func (connh *connectionHandler) Send(msg *screensv1.ServerMessages) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return connh.Write(data)
}

func (connh *connectionHandler) setupConnection(w http.ResponseWriter) {
	defer connh.close(nil)
	connh.logger.Info("New websocket connection established")

	old, loaded := connh.handler.Listeners.Swap(connh.id, connh)
	if loaded {
		connh.logger.Warn("A connection with this id already exists, closing old connection", "old_id", old.(*connectionHandler).id)
		old.(*connectionHandler).close(errors.New("replaced by new connection"))
	}

	defer func() {
		del := connh.handler.Listeners.CompareAndDelete(connh.id, connh)
		if !del {
			connh.logger.Warn("Could not remove listener from screen handler, I was replaced?")
		}
	}()

	readerType, reader, err := connh.conn.Reader(connh.ctx)
	if err != nil {
		connh.logger.Error("Could not open websocket reader", "error", err)
		http.Error(w, "Could not open websocket reader: "+err.Error(), http.StatusInternalServerError)

		return
	}
	if readerType != websocket.MessageBinary {
		connh.logger.Error("Unsupported websocket message type", "type", readerType)
		http.Error(w, "Unsupported websocket message type", http.StatusBadRequest)

		return
	}

	connh.readLoop(reader)
}

func (connh *connectionHandler) readLoop(reader io.Reader) {
	for {
		// Read
		message, err := io.ReadAll(reader)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				connh.logger.Info("Websocket connection closed")

				return
			}
			var ws websocket.CloseError
			if errors.As(err, &ws) {
				if ws.Code == websocket.StatusNormalClosure || ws.Code == websocket.StatusGoingAway {
					connh.logger.Info("Websocket connection closed normally", "code", ws.Code, "reason", ws.Reason)
					connh.close(nil)

					return
				}

				connh.logger.Info("Websocket connection closed by client", "code", ws.Code, "reason", ws.Reason)
				connh.close(ws)

				return
			}
			connh.logger.Error("Could not read websocket message", "error", err)

			return
		}
		// Parse
		var messageProto screensv1.ClientMessages
		err = proto.Unmarshal(message, &messageProto)
		if err != nil {
			connh.logger.Error("Could not parse websocket message", "error", err)
			connh.close(err)

			return
		}

		// Handle
		err = connh.handler.HandleMessages(connh.ctx, connh, &messageProto)
		if err != nil {
			connh.logger.Error("Could not handle websocket message", "error", err)
			connh.close(err)

			return
		}
	}
}

func (connh *connectionHandler) close(err error) {
	connh.logger.Debug("Closing connection...", "error", err)

	// Close the connection with the appropriate status code or normal closure.
	if err != nil {
		connh.logger.Error("Closing connection due to error", "error", err)
		connh.closeWith(websocket.StatusInternalError, "connection closed by server due to error: "+err.Error())

		return
	}

	connh.closeWith(websocket.StatusNormalClosure, "connection closed by server")
}

func (connh *connectionHandler) closeWith(code websocket.StatusCode, reason string) {
	crr := connh.conn.Close(code, reason)
	if errors.Is(crr, net.ErrClosed) { // Already closed, ignore.
		return
	}
	if crr != nil {
		connh.logger.Error("Could not close websocket connection", "error", crr)
	}
}
