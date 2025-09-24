package api

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/DaanV2/mechanus/server/mechanus/screens"
	"github.com/DaanV2/mechanus/server/pkg/authentication"
	"github.com/DaanV2/mechanus/server/pkg/authentication/roles"
	screensv1 "github.com/DaanV2/mechanus/server/pkg/grpc/gen/screens/v1"
	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrWebsocketNotFound is returned when a websocket connection could not be found.
	ErrWebsocketNotFound = errors.New("websocket: connection not found")
)

type DeviceType int

const (
	DeviceTypeUnknown DeviceType = iota
	DeviceTypeUser
	DeviceTypeDevice
)

var (
	_ http.Handler = &WebsocketHandler{}
)

type WebsocketInfo interface {
	io.Writer
	GetInfo() *ConnectionInfo
	Context() context.Context
}

type ConnectionInfo struct {
	Token string
	ID    string
	Roles []string
	Type  DeviceType
}

type WebsocketHandler struct {
	logger logging.Enriched
	config WebsocketConfig

	screenManager     *screens.ScreenManager
	jwt_authenticator *authentication.JWTService
}

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

func NewWebsocketRouter(handler *WebsocketHandler) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/api/v1/screen/{screenid}/{id}", handler) // Placeholder, will be handled by websocket handler.
	return router
}

func NewWebsocketHandler(handler *screens.ScreenManager, jwt_authenticator *authentication.JWTService, config WebsocketConfig) *WebsocketHandler {
	return &WebsocketHandler{
		logger:            logging.Enriched{}.WithPrefix("websocket"),
		config:            config,
		screenManager:     handler,
		jwt_authenticator: jwt_authenticator,
	}
}

func (handler *WebsocketHandler) acceptOptions() *websocket.AcceptOptions {
	return nil // TODO: Fill in options connected via config.
}

// ServeHTTP implements http.Handler.
func (handler *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	screenid := r.PathValue("screenid")
	if id == "" || screenid == "" {
		http.Error(w, "Missing id or type in path", http.StatusBadRequest)
		return
	}
	screenHandler, ok := handler.screenManager.Get(screenid)
	if !ok {
		http.Error(w, "Screen not found", http.StatusNotFound)
		return
	}
	//TODO: check that screenHandler is allowed to be access by this requester

	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, connCtx := handler.logger.
		With("id", id, "screen", screenid).
		FromUpdate(connCtx)
	r = r.WithContext(connCtx)

	auth, err := handler.authenticate(r)
	if err != nil {
		http.Error(w, "Could not authenticate: "+err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, handler.acceptOptions())
	if err != nil {
		logger.Error("Could not open websocket", "error", err)
		return
	}

	h := connectionHandler{
		id:       id,
		screenid: screenid,
		ctx:      connCtx,
		closeFn:  cancel,
		conn:     conn,
		logger:   logger,
		info:     auth,
		handler:  screenHandler,
	}

	h.setupConnection(w)
}

func (handler *WebsocketHandler) authenticate(r *http.Request) (*ConnectionInfo, error) {
	token := r.Header.Get("Authorization")

	// Bearer token are only supported for users for now.
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
		t, err := handler.jwt_authenticator.Validate(r.Context(), token)
		if err != nil {
			handler.logger.From(r.Context()).Error("Could not validate JWT", "error", err)
			return nil, err
		}
		c, ok := authentication.GetClaims(t.Claims)
		if !ok {
			handler.logger.From(r.Context()).Error("Could not get claims from JWT")
			return nil, errors.New("could not get claims from JWT")
		}

		return &ConnectionInfo{
			Token: token,
			ID:    c.User.ID,
			Roles: c.User.Roles,
			Type:  DeviceTypeUser,
		}, nil
	}

	return &ConnectionInfo{
		Token: token, // TODO: check token for device and device id.
		ID:    r.PathValue("id"),
		Roles: []string{roles.Device.String()},
		Type:  DeviceTypeDevice,
	}, nil
}

// Context implements WebsocketInfo.
func (connh *connectionHandler) Context() context.Context {
	return connh.ctx
}

// GetInfo implements WebsocketInfo.
func (connh *connectionHandler) GetInfo() *ConnectionInfo {
	return connh.info
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
		connh.conn.Close(websocket.StatusInternalError, "connection closed by server due to error: "+err.Error())
		return
	}

	crr := connh.conn.Close(websocket.StatusNormalClosure, "connection closed by server")
	if errors.Is(crr, net.ErrClosed) { // Already closed, ignore.
		return
	}
	if crr != nil {
		connh.logger.Error("Could not close websocket connection", "error", crr)
	}
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
