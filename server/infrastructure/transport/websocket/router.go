package websocket

import "net/http"

func NewWebsocketRouter(handler *WebsocketHandler) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/api/v1/screen/{screenid}/{id}", handler) // Placeholder, will be handled by websocket handler.

	return router
}
