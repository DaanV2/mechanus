# Communication & Protocols

Mechanus uses three transport mechanisms, all running on the same HTTP/2 listener.

## ConnectRPC (gRPC)

Used for request/response flows: login, user management, campaign CRUD, scene control.

**Proto definitions:** `server/proto/` compiled with [Buf](https://github.com/bufbuild/buf). Generated Go code (server) and TypeScript stubs (client, `src/proto/`) are both produced from the same sources.

Current services:

| Service           | Proto package  | Purpose                                                                     |
| ----------------- | -------------- | --------------------------------------------------------------------------- |
| `LoginService`    | `users/v1`     | Authenticate and issue JWTs                                                 |
| `UserService`     | `users/v1`     | User CRUD, admin list / password reset                                      |
| `CampaignService` | `campaigns/v1` | Create, get and list campaigns (admins see all, other users only their own) |
| `SceneService`    | `scenes/v1`    | GM scene control: list scenes, activate a scene, switch variant, get the active scene (any logged-in user) |

ConnectRPC is called from the browser using either the Connect protocol or gRPC-web over HTTP/2. A gRPC reflection endpoint and gRPC health endpoint are available for tooling and infrastructure probes.

**Client:** `client/src/lib/api/client.ts` sets up the transport; service-specific files (`users_v1.ts`, `campaigns_v1.ts`, etc.) wrap the generated stubs.

**Server:** `infrastructure/transport/routers` assembles the route table. gRPC interceptors (logging, auth, telemetry) are wired in `infrastructure/transport/grpc`.

## WebSocket — screen protocol

Used for the continuous bidirectional screen state stream. Each screen view (GM, player, device/TV) maintains a WebSocket connection and receives server-pushed state updates.

### Connection URL

```text
<ws|wss>://<host>[:<port>]/api/v1/screen/{screenId}/{connectionId}
```

**screenId** — the screen to show, which is the **device ID** of a screen device: there is one screen per screen device, and `ScreenManager.GetOrCreate` creates it on the first connection (or the first scene instruction) for that ID. What it shows is decided by the active scene through the room's target bindings — see [scenes.md](./scenes.md#screens--views). The device view (`/views/devices/`) uses its persisted device ID (`localStorage` `device-id`, overridable with `?device=<id>`).

**connectionId** — the connecting user ID or device ID.

### Authentication

Authentication is optional when connecting and enforced per route: the handler resolves the connection to a user, a paired device or an anonymous device, and the route's policy (`anonymous`, `device`, `user`, `role:<role>`) decides whether it may proceed. An invalid credential is rejected rather than treated as anonymous. See [authentication.md](./authentication.md#device-authentication) for the policies and the device pairing flow.

| Client type      | Credential                                                           |
| ---------------- | -------------------------------------------------------------------- |
| Player / GM      | User JWT, `Authorization: Bearer <token>` (browsers: see below)      |
| Paired device    | Device token from pairing, `Authorization: Device <token>` — planned |
| Anonymous device | None; only allowed on `anonymous` routes such as pairing — planned   |

**Current state**: only the `Bearer` path is implemented, and there are no route policies yet — any connection without a `Bearer` JWT is accepted as an unauthenticated device.

### Connection headers

Browsers cannot set headers on `new WebSocket(...)`, so every header that a browser client needs also has a query parameter fallback; the header wins when both are present.

| Header            | Query fallback | Description                                                                                |
| ----------------- | -------------- | ------------------------------------------------------------------------------------------ |
| `Authorization`   | see below      | `Bearer <jwt>` (user) or `Device <token>` (paired device); only `Bearer` is read today     |
| `X-Device-ID`     | `deviceId`     | Stable device ID, stored in `screens.ConnectionInfo.DeviceID` and on the connection's logger |
| `X-Device-Type`   | `deviceType`   | Hardware (`screen`, `table`, `tablet`, `phone`, `desktop`, `other`; unknown → `other`), stored in `ConnectionInfo.Hardware`. Metadata only |
| `X-Connection-ID` | —              | Planned: a UUID for tracking/telemetry per connection — not yet read by the handler        |

#### Browser credentials

A browser can't send `Authorization`, so a user JWT is also read from two browser-friendly places. `authentication.BearerFromHeaders` checks, in order:

1. `Authorization: Bearer <jwt>`
2. A `Sec-WebSocket-Protocol` entry `mechanus.bearer.<jwt>`. The client's `WebsocketHandler` opens the socket with `new WebSocket(url, ["mechanus.v1", "mechanus.bearer.<jwt>"])`; the server always selects `mechanus.v1`, so the token is never echoed back.
3. The `access-token` cookie, holding the URL-encoded `Bearer <jwt>` the client stores after login. Browsers send it on the upgrade request automatically.

A query parameter is deliberately not supported: URLs end up in logs and history. Cookie-based auth is safe from cross-site WebSocket hijacking only because `coder/websocket` rejects an `Origin` that doesn't match the `Host` by default — keep that check if `acceptOptions` ever sets `OriginPatterns`.

The handler authenticates before anything else and only creates the screen once the WebSocket is accepted, so failed or unauthenticated requests leave nothing behind.

### Message framing

Messages are proto3-encoded (`screens/v1`). The server sends `ServerMessages` (a batch of `ServerMessage`); the client sends `ClientMessages` (a batch of `ClientMessage`).

Request/response correlation: if a `ClientMessage` carries a non-empty `id` field, the server echoes it in all `ServerMessage` responses for that request.

### Supported client messages

| Message type          | Description                                                |
| --------------------- | ---------------------------------------------------------- |
| `Ping`                | Keep-alive / latency check                                 |
| `InitialSetupRequest` | Sent on connect; server responds with initial screen state, plus the splash screen while the screen has not been instructed (view `blank`) |

_(Additional messages for pan, zoom, token moves, etc. are not implemented yet.)_

### Server messages

| Message                                | Description                                                                                |
| -------------------------------------- | ------------------------------------------------------------------------------------------ |
| `InitialSetupResponse`                 | Answer to `InitialSetupRequest`, carries the screen's current `ScreenState`                |
| `ScreenState`                          | Full snapshot; the client replaces everything it renders. Sent by `ScreenHandler.SetState` |
| `EntityUpdate`                         | Add or replace one entity in a layer (token moved, revealed, spawned)                      |
| `EntityRemove`                         | Remove one entity from a layer                                                             |
| `Camera` (`camera_update`)             | Change the visible part of the view (pan / zoom)                                           |
| `SplashScreen`                         | Show / hide the title card                                                                 |
| `SceneChange` / `SceneChangeCompleted` | Scene load started / finished                                                              |
| `Ping`                                 | Broadcast of a client ping                                                                 |

`ScreenState` (`screens/v1/state.proto`) is what the Pixi.js renderer draws:

| Field           | Content                                                                                                                                                   |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `view`          | `ViewKind` (battlemap, scenery, infoboard, custom, splash, blank), its `Asset`, `component`, `params` and `overlay`s — mirrors `scenes.ScreenInstruction` |
| `layer`         | Ordered layers (bottom first) of `Entity`: id, name, `EntityKind`, asset, position / size in grid units, rotation, hidden                                 |
| `grid`          | Offset, cell size (in pixels of the view's asset), whether lines are shown and their colour                                                               |
| `camera`        | Center (grid units) and zoom                                                                                                                              |
| `environment`   | In-game time, background colour, light intensity                                                                                                          |
| `splash_screen` | Optional title card                                                                                                                                       |

The server-side model is `engine/screens/state`; `ScreenState.ToProto()` converts it. Delta messages (`EntityUpdate`, `EntityRemove`, `camera_update`) are defined but nothing sends them yet. Fog of war and walls are not in the model — they wait on the battle map format (see [scenes.md](./scenes.md#screens--views)).

### Server-side flow

`infrastructure/transport/http.WebsocketSplitter` intercepts HTTP requests with `Upgrade: websocket` before they reach the ConnectRPC router, handing them to `infrastructure/transport/websocket.WebsocketHandler`.

The handler authenticates the connection (resolving its identity and checking the route's policy — policies are planned), resolves the screen ID (the device ID), and registers the connection with `engine/screens.ScreenManager`. The `ScreenHandler` for that screen ID manages the listener set and broadcasts state to all connected clients for that screen.

### Scene assets

Scene assets (maps, images, video) are served by `infrastructure/assets.Store` under `/assets/`: scene assets from `<app config dir>/scenes/<scene id>/...` (referenced relative to the scene folder, `maps/cave.webp`), shared assets from `<app config dir>/shared/...` (referenced as `shared:/monsters/goblin.png`). The screen driver resolves references to URLs with a content hash (`?v=<hash>`); versioned requests are served `Cache-Control: immutable`. Only files inside a scene folder or the shared folder are served, never scene configs (`scenes/<id>.dat`) or other config files.

## Static file serving

The compiled SvelteKit SPA is served as static files from the path configured via `--server.static-folder` (default: `./static`). Any request not matched by ConnectRPC or WebSocket routes falls through to this file server. This means the Go server is the only process needed in production — no separate frontend server required.
