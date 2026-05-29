# Communication & Protocols

Mechanus uses three transport mechanisms, all running on the same HTTP/2 listener.

## ConnectRPC (gRPC)

Used for request/response flows: login, user management, campaign CRUD.

**Proto definitions:** `server/proto/` compiled with [Buf](https://github.com/bufbuild/buf). Generated Go code (server) and TypeScript stubs (client, `src/proto/`) are both produced from the same sources.

Current services:

| Service          | Proto package  | Purpose                         |
| ---------------- | -------------- | ------------------------------- |
| `LoginService`   | `users/v1`     | Authenticate and issue JWTs     |
| `UserService`    | `users/v1`     | User CRUD                       |
| _(campaigns/v1)_ | `campaigns/v1` | Campaign management _(planned)_ |

ConnectRPC is called from the browser using either the Connect protocol or gRPC-web over HTTP/2. A gRPC reflection endpoint and gRPC health endpoint are available for tooling and infrastructure probes.

**Client:** `client/src/lib/api/client.ts` sets up the transport; service-specific files (`users_v1.ts`, etc.) wrap the generated stubs.

**Server:** `infrastructure/transport/routers` assembles the route table. gRPC interceptors (logging, auth, telemetry) are wired in `infrastructure/transport/grpc`.

## WebSocket — screen protocol

Used for the continuous bidirectional screen state stream. Each screen view (GM, player, device/TV) maintains a WebSocket connection and receives server-pushed state updates.

### Connection URL

```text
<ws|wss>://<host>[:<port>]/api/v1/screen/{screenId}/{connectionId}
```

**screenId** — a UUID for a specific screen, or one of the named roles: `player`, `admin`, `viewer`.

**connectionId** — the connecting user ID or device ID.

### Authentication

| Client type | Credential                                                                 |
| ----------- | -------------------------------------------------------------------------- |
| Player / GM | JWT passed as a header or query parameter on the upgrade request           |
| Device      | API key; device ID sent via `X-Device-ID` header or `deviceid` query param |

See [authentication.md](./authentication.md) for how JWTs are issued and validated.

### Connection headers

| Header            | Description                                                               |
| ----------------- | ------------------------------------------------------------------------- |
| `X-Connection-ID` | UUID generated for this specific connection (used for tracking/telemetry) |
| `X-Device-ID`     | Device identifier for hardware screen clients                             |

### Message framing

Messages are proto3-encoded (`screens/v1`). The server sends `ServerMessages` (a batch of `ServerMessage`); the client sends `ClientMessages` (a batch of `ClientMessage`).

Request/response correlation: if a `ClientMessage` carries a non-empty `id` field, the server echoes it in all `ServerMessage` responses for that request.

### Supported client messages

| Message type          | Description                                                |
| --------------------- | ---------------------------------------------------------- |
| `Ping`                | Keep-alive / latency check                                 |
| `InitialSetupRequest` | Sent on connect; server responds with initial screen state |

_(Additional messages for pan, zoom, token moves, etc. are planned — see `TODO.md`.)_

### Server-side flow

`infrastructure/transport/http.WebsocketSplitter` intercepts HTTP requests with `Upgrade: websocket` before they reach the ConnectRPC router, handing them to `infrastructure/transport/websocket.WebsocketHandler`.

The handler authenticates the connection, resolves the screen ID, and registers the connection with `engine/screens.ScreenManager`. The `ScreenHandler` for that screen ID manages the listener set and broadcasts state to all connected clients for that screen.

## Static file serving

The compiled SvelteKit SPA is served as static files from the path configured via `--server.static-folder` (default: `./static`). Any request not matched by ConnectRPC or WebSocket routes falls through to this file server. This means the Go server is the only process needed in production — no separate frontend server required.
