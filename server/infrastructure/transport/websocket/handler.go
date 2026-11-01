package websocket

import (
	"errors"
	"net/http"
	"time"

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

	deviceID, hardware := deviceIdentity(r)
	logger, connCtx := handler.logger.
		With("id", id, "screen", screenid, "device_id", deviceID, "device_type", hardware).
		FromUpdate(r.Context())
	r = r.WithContext(connCtx)

	auth, err := handler.authenticate(r)
	if err != nil {
		http.Error(w, "Could not authenticate: "+err.Error(), http.StatusUnauthorized)

		return
	}
	auth.DeviceID, auth.Hardware = deviceID, hardware

	//TODO: check that screenHandler is allowed to be access by this requester

	conn, err := cwebsocket.Accept(w, r, handler.acceptOptions())
	if err != nil {
		logger.Error("Could not open websocket", "error", err)

		return
	}

	// One screen per screen device: the screen ID is the device ID. Only created once the connection is
	// accepted, so failed or unauthenticated requests don't leave screens behind.
	screenHandler := handler.screenManager.GetOrCreate(screenid)
	h := screens.NewScreenConn(id, screenid, connCtx, conn, auth)
	screenHandler.AddListener(connCtx, h)
	if auth.Type == devices.DeviceTypeUser {
		go handler.watchSession(h, auth)
	}

	// Wait until closed
	<-h.Context().Done()
}

// watchSession closes a user connection once its token expires, or its session (JTI) gets revoked,
// so a long lived connection doesn't outlive the credential it was opened with.
func (handler *WebsocketHandler) watchSession(h *screens.ScreenConn, auth *screens.ConnectionInfo) {
	ctx := h.Context()

	var expired <-chan time.Time
	if !auth.ExpiresAt.IsZero() {
		timer := time.NewTimer(time.Until(auth.ExpiresAt))
		defer timer.Stop()
		expired = timer.C
	}

	var check <-chan time.Time
	if handler.config.SessionCheckInterval > 0 {
		ticker := time.NewTicker(handler.config.SessionCheckInterval)
		defer ticker.Stop()
		check = ticker.C
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-expired:
			h.CloseWith(cwebsocket.StatusPolicyViolation, "token expired")

			return
		case <-check:
			if _, err := handler.jwt_authenticator.Validate(ctx, auth.Token); err != nil {
				handler.logger.From(ctx).Info("closing websocket, session no longer valid", "error", err)
				h.CloseWith(cwebsocket.StatusPolicyViolation, "session no longer valid")

				return
			}
		}
	}
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

		info := &screens.ConnectionInfo{
			Token: token,
			ID:    c.User.ID,
			Roles: c.User.Roles,
			Type:  devices.DeviceTypeUser,
		}
		if c.ExpiresAt != nil {
			info.ExpiresAt = c.ExpiresAt.Time
		}

		return info, nil
	}

	return &screens.ConnectionInfo{
		Token: r.Header.Get("Authorization"), // TODO: check token for device and device id.
		ID:    r.PathValue("id"),
		Roles: []string{roles.Device.String()},
		Type:  devices.DeviceTypeDevice,
	}, nil
}

// Device identity headers, with query parameter fallbacks for browsers (which can't set headers on a WebSocket).
const (
	DeviceIDHeader   = "X-Device-Id"
	DeviceTypeHeader = "X-Device-Type"
	DeviceIDQuery    = "deviceId"
	DeviceTypeQuery  = "deviceType"
)

// deviceIdentity reads the device ID and hardware type a client reports. Both are optional.
func deviceIdentity(r *http.Request) (string, devices.HardwareType) {
	id := r.Header.Get(DeviceIDHeader)
	if id == "" {
		id = r.URL.Query().Get(DeviceIDQuery)
	}
	hardware := r.Header.Get(DeviceTypeHeader)
	if hardware == "" {
		hardware = r.URL.Query().Get(DeviceTypeQuery)
	}

	return id, devices.ParseHardwareType(hardware)
}
