package websocket

import (
	"errors"
	"net/http"

	cwebsocket "github.com/coder/websocket"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

var (
	// ErrWebsocketNotFound is returned when a websocket connection could not be found.
	ErrWebsocketNotFound = errors.New("websocket: connection not found")
)

var (
	_ http.Handler = &WebsocketHandler{}
)

type WebsocketHandler struct {
	logger logging.Enriched
	config WebsocketConfig

	screenManager     *screens.ScreenManager
	jwt_authenticator *authentication.JWTService
}

func NewWebsocketHandler(handler *screens.ScreenManager, jwt_authenticator *authentication.JWTService, config WebsocketConfig) *WebsocketHandler {
	return &WebsocketHandler{
		logger:            logging.Enriched{}.WithPrefix("websocket"),
		config:            config,
		screenManager:     handler,
		jwt_authenticator: jwt_authenticator,
	}
}

func (handler *WebsocketHandler) acceptOptions() *cwebsocket.AcceptOptions {
	// TODO: Fill in options connected via config.
	return &cwebsocket.AcceptOptions{
		// A browser that offers subprotocols (to pass its token) fails the connection unless the server selects one.
		Subprotocols: []string{authentication.WebsocketSubprotocol},
	}
}

// ServeHTTP implements http.Handler.
func (handler *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	screenid := r.PathValue("screenid")
	if id == "" || screenid == "" {
		http.Error(w, "Missing id or type in path", http.StatusBadRequest)

		return
	}

	logger, connCtx := handler.logger.
		With("id", id, "screen", screenid).
		FromUpdate(r.Context())
	r = r.WithContext(connCtx)

	// Authenticate before looking up the screen, so unauthenticated requests can't probe which screens exist
	auth, err := handler.authenticate(r)
	if err != nil {
		http.Error(w, "Could not authenticate: "+err.Error(), http.StatusUnauthorized)

		return
	}

	screenHandler, ok := handler.screenManager.Get(screenid)
	if !ok {
		http.Error(w, "Screen not found", http.StatusNotFound)

		return
	}

	//TODO: check that screenHandler is allowed to be access by this requester

	conn, err := cwebsocket.Accept(w, r, handler.acceptOptions())
	if err != nil {
		logger.Error("Could not open websocket", "error", err)

		return
	}

	h := screens.NewScreenConn(id, screenid, connCtx, conn, auth)
	screenHandler.AddListener(connCtx, h)

	// Wait until closed
	<-h.Context().Done()
}

func (handler *WebsocketHandler) authenticate(r *http.Request) (*screens.ConnectionInfo, error) {
	// Bearer token are only supported for users for now.
	// Browsers can't set the Authorization header, so they pass it as a subprotocol or cookie instead.
	if token := authentication.BearerFromHeaders(r.Header); token != "" {
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
		if !c.HasScope(authentication.ScopeAll) {
			handler.logger.From(r.Context()).Error("JWT scope does not allow screen access", "scope", c.Scope)

			return nil, errors.New("JWT scope does not allow screen access")
		}

		return &screens.ConnectionInfo{
			Token: token,
			ID:    c.User.ID,
			Roles: c.User.Roles,
			Type:  devices.DeviceTypeUser,
		}, nil
	}

	return &screens.ConnectionInfo{
		Token: r.Header.Get("Authorization"), // TODO: check token for device and device id.
		ID:    r.PathValue("id"),
		Roles: []string{roles.Device.String()},
		Type:  devices.DeviceTypeDevice,
	}, nil
}
