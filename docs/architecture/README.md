# Architecture

Mechanus is a server-driven UI system for local TTRPG setups. A single Go server manages state, authentication, and asset delivery. Browser-based clients connect to it and render their assigned view (GM dashboard, player screen, table/TV battle map, etc.).

## Documents

| Doc                                      | What it covers                                                                        |
| ---------------------------------------- | ------------------------------------------------------------------------------------- |
| [server.md](./server.md)                 | Go server — Clean Architecture layers, startup/shutdown lifecycle, HTTP handler stack |
| [client.md](./client.md)                 | SvelteKit + Pixi.js client — view types, rendering, toolchain                         |
| [communication.md](./communication.md)   | ConnectRPC/gRPC services, WebSocket screen protocol, static file serving              |
| [authentication.md](./authentication.md) | JWT issuance & validation, RSA key management, JTI revocation, device auth            |
| [database.md](./database.md)             | Supported backends, configuration, schema (all tables)                                |
| [mdns.md](./mdns.md)                     | LAN service discovery via Multicast DNS                                               |

## High-level overview

```ascii
┌─────────────────────────────────────────────────────────────────┐
│                          Browser clients                        │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │  GM view     │  │  Player view │  │  Device / TV view     │  │
│  │  (SvelteKit) │  │  (SvelteKit) │  │  (Pixi.js fullscreen) │  │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬───────────┘  │
└─────────┼─────────────────┼──────────────────────┼──────────────┘
          │  ConnectRPC     │  WebSocket (screen)  │
          ▼                 ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Mechanus server (Go)                    │
│                                                                 │
│  HTTP/2 router ──► ConnectRPC handlers ──► application layer    │
│       └──────────► WebSocket handler  ──► engine (screens)      │
│       └──────────► Static file server                           │
│                                                                 │
│  Infrastructure: SQLite (GORM), JWT auth, OpenTelemetry, mDNS   │
└─────────────────────────────────────────────────────────────────┘
```

## Key design decisions

**Server-driven UI** — the server owns all scene/game state and pushes it to clients over WebSocket. Clients are thin renderers.

**Single binary** — the Go server serves both the compiled SvelteKit SPA and all API endpoints from one process. No separate frontend server is needed in production.

**ConnectRPC for RPC, WebSocket for streaming** — ConnectRPC (gRPC over HTTP/2) handles request/response flows (auth, user management). WebSocket handles the continuous bidirectional screen state stream.

**Clean Architecture layers** — the server is split into `engine` (domain), `application` (use cases), `infrastructure` (adapters), and `components` (wiring). See [server.md](./server.md).

**Proto3 as the contract** — all client↔server message shapes are defined in `.proto` files under `server/proto/`. Generated code lives alongside them via Buf.
