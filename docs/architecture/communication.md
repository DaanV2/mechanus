# Communication & Protocols

Mechanus uses three transport mechanisms, all running on the same HTTP/2 listener.

## ConnectRPC (gRPC)

Used for request/response flows: login, user management, campaign CRUD.

**Proto definitions:** `server/proto/` compiled with [Buf](https://github.com/bufbuild/buf). Generated Go code (server) and TypeScript stubs (client, `src/proto/`) are both produced from the same sources.

Current services:

| Service           | Proto package  | Purpose                                                                     |
| ----------------- | -------------- | --------------------------------------------------------------------------- |
| `LoginService`    | `users/v1`     | Authenticate and issue JWTs                                                 |
| `UserService`     | `users/v1`     | User CRUD, admin list / password reset                                      |
| `CampaignService` | `campaigns/v1` | Create, get and list campaigns (admins see all, other users only their own) |

ConnectRPC is called from the browser using either the Connect protocol or gRPC-web over HTTP/2. A gRPC reflection endpoint and gRPC health endpoint are available for tooling and infrastructure probes.

**Client:** `client/src/lib/api/client.ts` sets up the transport; service-specific files (`users_v1.ts`, `campaigns_v1.ts`, etc.) wrap the generated stubs.

**Server:** `infrastructure/transport/routers` assembles the route table. gRPC interceptors (logging, auth, telemetry) are wired in `infrastructure/transport/grpc`.

## WebSocket — screen protocol

Used for the continuous bidirectional screen state stream. Each screen view (GM, player, device/TV) maintains a WebSocket connection and receives server-pushed state updates.

### Connection URL

```text
<ws|wss>://<host>[:<port>]/api/v1/screen/{screenId}/{connectionId}
```

**screenId** — the ID of a screen already registered with `engine/screens.ScreenManager`. Resolving a device to its screen is not implemented yet; the intended design is one screen per screen device, driven by the active scene through the room's target bindings — see [scenes.md](./scenes.md#screens--views).

**connectionId** — the connecting user ID or device ID.

### Authentication

| Client type | Credential                                                                                                                                                              |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Player / GM | JWT passed as `Authorization: Bearer <token>` on the upgrade request                                                                                                     |
| Device      | **Not implemented yet.** Any connection without a `Bearer` JWT is currently accepted as an unauthenticated device — see [authentication.md](./authentication.md#device-authentication). |

See [authentication.md](./authentication.md) for how JWTs are issued and validated.

### Connection headers

| Header            | Description                                                                                 |
| ----------------- | --------------------------------------------------------------------------------------------- |
| `Authorization`   | `Bearer <jwt>` for player/GM connections — the only header currently read by the handler       |
| `X-Connection-ID` | Planned: a UUID for tracking/telemetry per connection — not yet read by the handler             |
| `X-Device-ID`     | Planned: device identifier for hardware screen clients — not yet read by the handler            |

### Message framing

Messages are proto3-encoded (`screens/v1`). The server sends `ServerMessages` (a batch of `ServerMessage`); the client sends `ClientMessages` (a batch of `ClientMessage`).

Request/response correlation: if a `ClientMessage` carries a non-empty `id` field, the server echoes it in all `ServerMessage` responses for that request.

### Supported client messages

| Message type          | Description                                                |
| --------------------- | ---------------------------------------------------------- |
| `Ping`                | Keep-alive / latency check                                 |
| `InitialSetupRequest` | Sent on connect; server responds with initial screen state |

_(Additional messages for pan, zoom, token moves, etc. are not implemented yet.)_

### Server messages

| Message                  | Description                                                                                          |
| ------------------------ | ---------------------------------------------------------------------------------------------------- |
| `InitialSetupResponse`   | Answer to `InitialSetupRequest`, carries the screen's current `ScreenState`                          |
| `ScreenState`            | Full snapshot; the client replaces everything it renders. Sent by `ScreenHandler.SetState`           |
| `EntityUpdate`           | Add or replace one entity in a layer (token moved, revealed, spawned)                                 |
| `EntityRemove`           | Remove one entity from a layer                                                                        |
| `Camera` (`camera_update`) | Change the visible part of the view (pan / zoom)                                                   |
| `SplashScreen`           | Show / hide the title card                                                                            |
| `SceneChange` / `SceneChangeCompleted` | Scene load started / finished                                                          |
| `Ping`                   | Broadcast of a client ping                                                                            |

`ScreenState` (`screens/v1/state.proto`) is what the Pixi.js renderer draws:

| Field          | Content                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `view`         | `ViewKind` (battlemap, scenery, infoboard, custom, splash, blank), its `Asset`, `component`, `params` and `overlay`s — mirrors `scenes.ScreenInstruction` |
| `layer`        | Ordered layers (bottom first) of `Entity`: id, name, `EntityKind`, asset, position / size in grid units, rotation, hidden |
| `grid`         | Offset, cell size (in pixels of the view's asset), whether lines are shown and their colour                |
| `camera`       | Center (grid units) and zoom                                                                               |
| `environment`  | In-game time, background colour, light intensity                                                           |
| `splash_screen`| Optional title card                                                                                        |

The server-side model is `engine/screens/state`; `ScreenState.ToProto()` converts it. Delta messages (`EntityUpdate`, `EntityRemove`, `camera_update`) are defined but nothing sends them yet. Fog of war and walls are not in the model — they wait on the battle map format (see [scenes.md](./scenes.md#screens--views)).

### Server-side flow

`infrastructure/transport/http.WebsocketSplitter` intercepts HTTP requests with `Upgrade: websocket` before they reach the ConnectRPC router, handing them to `infrastructure/transport/websocket.WebsocketHandler`.

The handler authenticates the connection, resolves the screen ID, and registers the connection with `engine/screens.ScreenManager`. The `ScreenHandler` for that screen ID manages the listener set and broadcasts state to all connected clients for that screen.

## Static file serving

The compiled SvelteKit SPA is served as static files from the path configured via `--server.static-folder` (default: `./static`). Any request not matched by ConnectRPC or WebSocket routes falls through to this file server. This means the Go server is the only process needed in production — no separate frontend server required.
