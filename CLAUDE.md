# Mechanus — AI Context

Mechanus is a server-driven UI system for local TTRPG (tabletop RPG) setups. A Go server manages all game state and pushes it to browser clients (GM dashboard, player screens, TV/table battle map display) over ConnectRPC (gRPC) and WebSocket.

## Repo layout

```
mechanus/
├── server/       Go server (main codebase)
├── client/       SvelteKit + Pixi.js browser client
├── tests/        Integration tests (Playwright)
├── docs/         Documentation
│   └── architecture/   Architecture docs — read these first
├── buf.yaml      Proto compiler config (Buf)
└── docker-compose.yaml
```

## Architecture

Read `docs/architecture/README.md` for a full overview. Short version:

- **Server** follows Clean Architecture: `engine` (domain) → `application` (use cases) → `infrastructure` (adapters). `components` is the DI root that wires everything.
- **Client** is a SvelteKit SPA with three view types: GM, player, device/TV. Pixi.js handles 2D canvas rendering. ConnectRPC + WebSocket for server communication.
- **Proto3** is the contract between client and server. Files live in `server/proto/`, generated with Buf.
- **Single binary in production**: the Go server serves both the API and the compiled SvelteKit SPA.

Key architecture docs:
- `docs/architecture/server.md` — layers, lifecycle, HTTP handler stack
- `docs/architecture/communication.md` — ConnectRPC services, WebSocket screen protocol
- `docs/architecture/authentication.md` — JWT / JTI / RSA key flow
- `docs/architecture/database.md` — schema, backends, config
- `docs/architecture/mdns.md` — LAN discovery

## Development

Start everything with Docker (recommended):
```bash
docker compose watch   # auto-rebuilds on change
```

Or run server directly:
```bash
cd server && go run main.go server
```

See `docs/development.md` for the full local dev guide.

## Common commands

### Root
```bash
just proto          # regenerate Go + TS code from .proto files
just checks         # proto gen + format + lint + docker build
just server         # build server binary
just dev            # docker compose watch
```

### Server (`cd server`)
```bash
just build          # go build ./...
just test           # run all tests (Ginkgo, with coverage)
just lint           # golangci-lint --fix
just format         # go fmt
just checks         # build + test + format + lint + docs
just start-server   # go run main.go server
```

### Client (`cd client`)
```bash
npm install
npm run dev         # vite dev server
npm run build       # production build → output served by Go server
npm run lint
npm run check       # svelte-check + tsc
```

## Adding a new gRPC service

1. Add `.proto` file in `server/proto/<domain>/v1/`
2. Run `just proto` from the repo root to generate Go + TypeScript stubs
3. Implement the service interface in `server/application/` or `server/infrastructure/`
4. Register the handler in `server/components/servers.go` (`CreateRouter`)
5. Add client wrapper in `client/src/lib/api/`

## Adding a new database model

1. Create the struct in `server/infrastructure/persistence/models/`
2. Embed `models.Model` for the standard ID/timestamps/soft-delete
3. Add the model to the `ApplyMigrations` call in `server/components/database.go`
4. Add a repository in `server/infrastructure/persistence/repositories/`

## Key conventions

- **Layer boundaries**: `engine` must not import `infrastructure`. `application` may import `engine` and define ports (interfaces). `infrastructure` implements those ports. `components` is the only place that crosses all layers.
- **Config**: use Viper config sets (`infrastructure/config`). Add flags to the relevant `*ConfigSet` and register them in the appropriate `cmd/*.go` file.
- **Error handling**: wrap errors with `fmt.Errorf("context: %w", err)`. Use `errors.Join` to combine multiple errors.
- **Logging**: use `infrastructure/logging` (Charm.sh). Pass `ctx` through and use `logging.From(ctx)` to get a logger with trace context.
- **Tests**: Ginkgo + Gomega for Go unit/component tests. Playwright for integration/browser tests in `tests/`.
- **Proto style**: follow Buf lint rules. Run `just proto-lint` before committing proto changes.

## Current state & what needs work

See `TODO.md` for the full backlog. High-priority gaps:

- **Scene management** — `engine/scenes` has the structure but `ScreenIDForRole` and `ScreenIDForDevice` are unimplemented stubs
- **Battle map rendering pipeline** — server needs to process `.dd2vtt` / `.dungeondraft_map` files into screen state
- **GM view** — client dashboard for controlling scenes and tokens
- **Player view** — limited-visibility perspective
- **WebSocket message handling** — only `Ping` and `InitialSetupRequest` are handled; pan, zoom, token moves are not yet wired up
- **Device registration** — devices can connect but assignment to screens is not persisted

## Environment variables

See `.env.example`. Key ones:

| Variable | Description |
| --- | --- |
| `WEB_PORT` | HTTP listen port (default `8080`) |
| `WEB_STATIC_FOLDER` | Path to compiled SvelteKit output |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `OTEL_ENABLED` | Enable OpenTelemetry tracing |
| `INITIALIZE_ADMIN_USERNAME` / `_PASSWORD` | Seed an admin user on first start |

## Testing

```bash
# Server unit tests
cd server && just test

# Integration tests (requires a running server)
cd tests && npx playwright test
```

Alternative Docker Compose setups for testing different DB backends:
- `docker-compose.sqlite.yaml`
- `docker-compose.postgres.yaml`
- `docker-compose.inmemory.yaml`
