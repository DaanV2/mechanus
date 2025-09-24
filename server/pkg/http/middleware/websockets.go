package http_middleware

import "net/http"

var _ http.Handler = &WebsocketSplitter{}

type WebsocketSplitter struct {
	websocket, other http.Handler
}

// NewWebsocketSplitter creates a new handler that splits between websocket requests and other requests.
func NewWebsocketSplitter(websocket, other http.Handler) *WebsocketSplitter {
	return &WebsocketSplitter{
		websocket, other,
	}
}

// ServeHTTP implements http.Handler.
func (w *WebsocketSplitter) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme == "ws" || req.URL.Scheme == "wss" {
		w.websocket.ServeHTTP(writer, req)

		return
	}

	w.other.ServeHTTP(writer, req)
}
