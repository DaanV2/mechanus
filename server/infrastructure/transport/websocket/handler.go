package websocket

import (
	"errors"
	"net/http"
	"time"

	cwebsocket "github.com/coder/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	mhttp "github.com/DaanV2/mechanus/server/infrastructure/transport/http"
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
	identities        *authentication.IdentityResolver
}

func NewWebsocketHandler(handler *screens.ScreenManager, jwt_authenticator *authentication.JWTService, config WebsocketConfig) *WebsocketHandler {
	return &WebsocketHandler{
		logger:            logging.Enriched{}.WithPrefix("websocket"),
		config:            config,
		screenManager:     handler,
		jwt_authenticator: jwt_authenticator,
		identities:        authentication.NewIdentityResolver(jwt_authenticator),
	}
}

// Route returns the handler guarded by the policy, the route's identity ends up on the request's context.
func (handler *WebsocketHandler) Route(policy authentication.Policy) http.Handler {
	return mhttp.RequirePolicy(handler.identities, policy, handler)
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

	identity := authentication.IdentityFromContext(r.Context())
	deviceID, hardware := deviceIdentity(r)
	// For a paired device the ID in its token is authoritative
	if identity.Kind == authentication.IdentityDevice {
		if deviceID != "" && deviceID != identity.ID {
			http.Error(w, "Device ID doesn't match the device token", http.StatusUnauthorized)

			return
		}
		deviceID = identity.ID
	}

	// The span comes from the HTTP tracing middleware and lasts as long as the connection.
	trace.SpanFromContext(r.Context()).SetAttributes(
		attribute.String("device.id", deviceID),
		attribute.String("device.type", string(hardware)),
		attribute.String("screen.id", screenid),
	)
	logger, connCtx := handler.logger.
		With("id", id, "screen", screenid, "device_id", deviceID, "device_type", hardware).
		FromUpdate(r.Context())
	r = r.WithContext(connCtx)

	auth := connectionInfo(identity)
	auth.DeviceID, auth.Hardware = deviceID, hardware

	if !auth.CanAccess(screenid) {
		logger.Warn("screen access denied", "connection_type", auth.Type)
		http.Error(w, "Not allowed to access this screen", http.StatusForbidden)

		return
	}

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
	if auth.Token != "" {
		go handler.watchSession(h, auth)
	}

	// Wait until closed
	<-h.Context().Done()
}

// watchSession closes a connection once its token expires, or its session (JTI) gets revoked,
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

// connectionInfo describes the connection of the identity.
func connectionInfo(identity *authentication.Identity) *screens.ConnectionInfo {
	info := &screens.ConnectionInfo{
		Token:     identity.Token,
		ID:        identity.ID,
		ExpiresAt: identity.ExpiresAt,
	}

	switch identity.Kind {
	case authentication.IdentityUser:
		info.Type = devices.DeviceTypeUser
		info.Roles = identity.Claims.User.Roles
	case authentication.IdentityDevice:
		info.Type = devices.DeviceTypeDevice
		info.Roles = []string{roles.Device.String()}
	case authentication.IdentityAnonymous:
		info.Type = devices.DeviceTypeUnknown
	}

	return info
}

// Device identity headers, with query parameter fallbacks for browsers (which can't set headers on a WebSocket).
const (
	DeviceIDHeader   = authentication.DeviceIDHeader
	DeviceTypeHeader = authentication.DeviceTypeHeader
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
