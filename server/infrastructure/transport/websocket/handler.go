package websocket

import (
	"errors"
	"net/http"
	"strings"

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
	logger, connCtx := handler.logger.
		With("id", id, "screen", screenid).
		FromUpdate(r.Context())
	r = r.WithContext(connCtx)

	auth, err := handler.authenticate(r)
	if err != nil {
		http.Error(w, "Could not authenticate: "+err.Error(), http.StatusUnauthorized)

		return
	}

	conn, err := cwebsocket.Accept(w, r, handler.acceptOptions())
	if err != nil {
		logger.Error("Could not open websocket", "error", err)

		return
	}

	h := screens.NewScreenConn(id, screenid, connCtx, conn, auth)
	screenHandler.AddListener(connCtx, h)

	// Wait until closed
	_ = <- h.Context().Done()
}

func (handler *WebsocketHandler) authenticate(r *http.Request) (*screens.ConnectionInfo, error) {
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

		return &screens.ConnectionInfo{
			Token: token,
			ID:    c.User.ID,
			Roles: c.User.Roles,
			Type:  devices.DeviceTypeUser,
		}, nil
	}

	return &screens.ConnectionInfo{
		Token: token, // TODO: check token for device and device id.
		ID:    r.PathValue("id"),
		Roles: []string{roles.Device.String()},
		Type:  devices.DeviceTypeDevice,
	}, nil
}
