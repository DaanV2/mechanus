# Web Server - API

The Go server is the single entry point for the application: it serves the compiled SvelteKit SPA, the ConnectRPC API, and the screen WebSocket, all from one `net/http` listener.

Default port is `8080` (`--server.port` / `SERVER_PORT`); see [configuration/config-file.md](../configuration/config-file.md) and [architecture/server.md](../architecture/server.md#http-handler-stack) for the full handler stack.

## Routes

| Path                       | Handled by                                              |
| --------------------------- | -------------------------------------------------------- |
| `/api/v1/screen/{screenId}/{id}` | WebSocket screen protocol — see [websocket.md](./websocket.md) |
| `/users.v1.LoginService/*`, `/users.v1.UserService/*` | ConnectRPC handlers (login, user management) |
| `/health`, `/healthz`      | HTTP health check                                         |
| `/ready`, `/readyz`        | HTTP readiness check                                      |
| `/` (anything else)        | Static file server for the compiled SvelteKit SPA (`--server.static.folder` / `SERVER_STATIC_FOLDER`, default `/web`) |

ConnectRPC also exposes gRPC reflection and gRPC health endpoints for tooling.

See [architecture/communication.md](../architecture/communication.md) for protocol details and [architecture/authentication.md](../architecture/authentication.md) for how requests are authenticated.
