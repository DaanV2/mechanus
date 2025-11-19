package screens

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	screensv1 "github.com/DaanV2/mechanus/server/proto/screens/v1"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

// ScreenConn represents a WebSocket connection to a screen.
type ScreenConn struct {
	id       string // user / device id
	screenid string // screen id
	connctx  context.Context
	cancelfn func()
	conn     *websocket.Conn
	logger   logging.Enriched
	info     *ConnectionInfo
}

// NewScreenConn creates a new screen connection with the provided parameters.
func NewScreenConn(id, screenid string, connctx context.Context, conn *websocket.Conn, info *ConnectionInfo) *ScreenConn {
	wctx, cancelfn := context.WithCancel(connctx)

	return &ScreenConn{
		id:       id,
		screenid: screenid,
		connctx:  wctx,
		cancelfn: cancelfn,
		conn:     conn,
		info:     info,
		logger: logging.Enriched{}.
			With("conn_id", id, "screen", screenid),
	}
}

// ConnectionInfo contains authentication and identification information for a connection.
type ConnectionInfo struct {
	Token string
	ID    string
	Roles []string
	Type  devices.DeviceType
}

// Context returns the connection's context.
func (sconn *ScreenConn) Context() context.Context {
	return sconn.connctx
}

// Send sends a server message to the client over the WebSocket connection.
func (sconn *ScreenConn) Send(ctx context.Context, msg *screensv1.ServerMessages) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return sconn.conn.Write(ctx, websocket.MessageBinary, data)
}

// Close handles the shutdown and message around closing connection,
// if the error is not nil it will log and report it to the listener with an [websocket.StatusInternalError]
// Or if passed as a [websocket.CloseError] then it grabs the code from the error
func (connh *ScreenConn) Close(err error) {
	connh.logger.Debug(connh.connctx, "Closing connection...", "error", err)

	// Close the connection with the appropriate status code or normal closure.
	if err != nil {
		code := websocket.CloseStatus(err)
		if code < 0 {
			code = websocket.StatusInternalError
		}

		connh.logger.Error(connh.connctx, "Closing connection due to error", "error", err)
		connh.CloseWith(code, "connection closed by server due to error: "+err.Error())

		return
	}

	connh.CloseWith(websocket.StatusNormalClosure, "connection closed by server")
}

func (connh *ScreenConn) startReadLoop(handleMessage func(ctx context.Context, connh *ScreenConn, msg *screensv1.ClientMessages) error) {
	logger := connh.logger.From(connh.connctx)
	readerType, reader, err := connh.conn.Reader(connh.connctx)
	if err != nil {
		logger.Error("Could not open websocket reader", "error", err)
		connh.Close(err)

		return
	}
	if readerType != websocket.MessageBinary {
		logger.Error("Unsupported websocket message type", "type", readerType)
		connh.Close(err)

		return
	}

	for {
		// Read
		message, err := io.ReadAll(reader)
		if err != nil {
			logger = logger.With("error", err)
			if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) {
				logger.Info("Websocket connection closed")

				return
			}
			code := websocket.CloseStatus(err)
			if code == websocket.StatusNormalClosure || code == websocket.StatusGoingAway {
				logger.Info("Websocket connection closed normally", "code", code)
				connh.Close(nil)

				return
			}
			if code > 0 {
				logger.Info("Websocket connection closed by client", "code", code)
			}

			logger.Error("Could not read websocket message", "error", err)
			connh.Close(err)

			return
		}

		// Parse
		var messageProto screensv1.ClientMessages
		err = proto.Unmarshal(message, &messageProto)
		if err != nil {
			logger.Error("Could not parse websocket message", "error", err)
			connh.Close(err)

			return
		}

		// Handle
		err = handleMessage(connh.connctx, connh, &messageProto)
		if err != nil {
			logger.Error("Could not handle websocket message", "error", err)
			connh.Close(err)

			return
		}
	}
}

// CloseWith closes the WebSocket connection with the specified status code and reason.
func (connh *ScreenConn) CloseWith(code websocket.StatusCode, reason string) {
	defer connh.cancelfn()

	crr := connh.conn.Close(code, reason)
	if errors.Is(crr, net.ErrClosed) { // Already closed, ignore.
		return
	}
	if crr != nil {
		connh.logger.From(connh.connctx).Error("Could not close websocket connection", "error", crr)
	}
}
