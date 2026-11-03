package websocket

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
)

func NewWebsocketRouter(handler *WebsocketHandler) *http.ServeMux {
	router := http.NewServeMux()
	// Room devices and users, the per-screen access check runs after this policy
	router.Handle("/api/v1/screen/{screenid}/{id}", handler.Route(authentication.PolicyDevice))

	return router
}
