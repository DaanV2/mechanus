package screens

import (
	"context"
	"errors"
	"net"
	"time"

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
			With("conn_id", id, "screen", screenid, "device_id", info.DeviceID, "device_type", info.Hardware),
	}
}

// ConnectionInfo contains authentication and identification information for a connection.
type ConnectionInfo struct {
	Token    string
	ID       string
	Roles    []string
	Type     devices.DeviceType
	DeviceID string               // Stable device ID reported by the client (X-Device-ID), empty when not given
	Hardware devices.HardwareType // Hardware reported by the client (X-Device-Type), metadata only
	// When the credential of the connection expires, zero if it doesn't.
	ExpiresAt time.Time
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
	// However the loop ends, the connection is over: release whoever waits on its context
	defer connh.cancelfn()

	logger := connh.logger.From(connh.connctx)

	for {
		// Read
		messageType, message, err := connh.conn.Read(connh.connctx)
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

		if messageType != websocket.MessageBinary {
			logger.Error("Unsupported websocket message type", "type", messageType)
			connh.CloseWith(websocket.StatusUnsupportedData, "only binary messages are supported")

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

// CanAccess reports whether the connection may open the screen with the given ID.
// Screens don't have owners yet, so for now: users may open every screen (GM and player views),
// a device only its own screen, whose ID is its device ID.
func (info *ConnectionInfo) CanAccess(screenID string) bool {
	switch info.Type {
	case devices.DeviceTypeUser:
		return true
	case devices.DeviceTypeDevice:
		return info.DeviceID != "" && info.DeviceID == screenID
	case devices.DeviceTypeUnknown:
	}

	return false
}
