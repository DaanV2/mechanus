# Server Architecture

The server is a single Go binary built around Clean Architecture. Code is split into four packages that correspond to the classic layers, plus a CLI entry point.

## Layers

```ascii
server/
├── cmd/              CLI entry points (Cobra commands)
├── engine/           Domain layer — core rules, no external deps
├── application/      Use-case layer — orchestrates domain + ports
├── components/       Composition root — wires everything together (DI container)
├── infrastructure/   Adapters — persistence, transport, auth, telemetry, …
└── pkg/              Shared utilities (no domain knowledge)
```

### engine — domain

Pure domain logic. Zero imports from infrastructure. Current domains:

- `engine/authentication` — JWT claims shape, JTI (token ID) revocation logic
- `engine/scenes` — maps roles/devices to screen IDs _(partially implemented)_
- `engine/screens` — `ScreenHandler` and `ScreenManager`; manages WebSocket listeners per screen, broadcasts state updates

### application — use cases

Orchestrates flows by combining engine rules with infrastructure ports:

- `UserService` — user CRUD, password handling
- `CampaignService` — campaign CRUD
- `Manager` — top-level coordinator _(stub)_

### infrastructure — adapters

All technical adapters. See dedicated sub-docs for detail:

| Package               | Responsibility                                                 | Doc                                                                       |
| --------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `authentication`      | JWT issuance & validation, RSA key management, JTI service     | [authentication.md](./authentication.md)                                  |
| `config`              | Viper-based config loading                                     | [../../configuration/config-file.md](../configuration/config-file.md)     |
| `health`              | `/health` and `/ready` HTTP endpoints                          | —                                                                         |
| `lifecycle`           | Ordered AfterInitialize / BeforeShutdown / AfterShutDown hooks | —                                                                         |
| `logging`             | Charm.sh structured logger + HTTP middleware                   | —                                                                         |
| `persistence`         | GORM + SQLite/Postgres/MySQL; auto-migrations; repositories    | [database.md](./database.md)                                              |
| `servers`             | `net/http` server wrapper with graceful shutdown               | —                                                                         |
| `storage`             | Blob / asset storage                                           | —                                                                         |
| `telemetry`           | OpenTelemetry setup; trace HTTP middleware                     | [../../configuration/opentelemetry.md](../configuration/opentelemetry.md) |
| `transport/cors`      | CORS middleware                                                | —                                                                         |
| `transport/grpc`      | gRPC interceptors                                              | [communication.md](./communication.md)                                    |
| `transport/http`      | WebSocket splitter                                             | [communication.md](./communication.md)                                    |
| `transport/mdns`      | mDNS LAN discovery                                             | [mdns.md](./mdns.md)                                                      |
| `transport/routers`   | Assembles all HTTP routes                                      | [communication.md](./communication.md)                                    |
| `transport/websocket` | WebSocket upgrade + screen state streaming                     | [communication.md](./communication.md)                                    |
| `vttrpg`              | VTTRPG file format parsers _(planned)_                         | —                                                                         |

### components — composition root

`components.BuildServer` is the single place where all layers are wired together. It reads config, constructs every infrastructure object, injects dependencies into application services, and returns a `ServerComponents` struct. Nothing else reaches across layer boundaries.

## Startup & shutdown lifecycle

```ascii
cmd/server.go
  └─► components.BuildServer(ctx)
        └─► cmpts.Components.AfterInitialize(ctx)   // DB migrations, key loading, …
  └─► server.Listen()                                // non-blocking HTTP/2 listener
  └─► <-ctx.Done()                                   // wait for SIGINT / SIGTERM
  └─► cmpts.Components.BeforeShutdown(ctx)           // drain / unregister
  └─► server.Shutdown(ctx)                           // graceful HTTP shutdown (1 min timeout)
  └─► cmpts.Components.AfterShutDown(ctx)
  └── cmpts.Components.ShutdownCleanup(ctx)
```

## HTTP handler stack

From outermost (first to receive a request) to innermost:

```ascii
Logging middleware
OpenTelemetry trace middleware
WebSocket splitter          ← upgrades go to WebSocket router; everything else falls through
CORS middleware
  ├── WebSocket router       (screen WebSocket connections)
  └── HTTP mux
        ├── ConnectRPC handlers  (Login, User, gRPC health, gRPC reflect)
        ├── /health, /healthz    (HTTP health)
        ├── /ready,  /readyz     (HTTP readiness)
        └── /                   (static file server — compiled SvelteKit SPA)
```
