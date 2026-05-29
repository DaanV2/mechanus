# Client Architecture

The client is a SvelteKit SPA compiled to static files and served by the Go server. It uses [Pixi.js](https://pixijs.com/) for 2D canvas rendering on screen views and [ConnectRPC](https://connectrpc.com/) for gRPC calls to the server.

## View types

There are three distinct views a browser client can render, determined server-side based on the screen ID assigned to the connection:

| View        | Route                | Purpose                                                |
| ----------- | -------------------- | ------------------------------------------------------ |
| GM view     | `/views/game-master` | Dashboard for controlling scenes, tokens, and maps     |
| Player view | `/views/players`     | Player perspective with limited visibility             |
| Device view | `/views/devices`     | Fullscreen Pixi.js battle map for a TV or table screen |

View assignment: when a client connects over WebSocket it sends an `InitialSetupRequest`; the server responds with the screen state for that screen ID. See [communication.md](./communication.md).

## Rendering

**SvelteKit + Svelte** for all non-canvas views (GM dashboard, campaign management, login, etc.). Styled with TailwindCSS and Flowbite-Svelte components.

**Pixi.js** for the device/TV view and any canvas-based 2D rendering. The entry point is `lib/2d/application.ts`. A composable layer system (background map, grid, tokens, fog of war, UI overlay) is planned under `lib/2d/components/`.

## Server communication

See [communication.md](./communication.md) for the full protocol details. From the client side:

- `lib/api/` — ConnectRPC transport and generated service clients (login, user management)
- `lib/networking/websocket.ts` — WebSocket connection lifecycle
- `lib/networking/server-events.ts` — handles incoming `ServerMessage` batches

Proto-generated TypeScript types in `src/proto/` are shared between both channels.

## Authentication

JWTs issued by the server are stored client-side and attached to ConnectRPC requests. For device connections an API key is used instead, passed as the `X-Device-ID` header or `deviceid` query parameter. See [authentication.md](./authentication.md).

## Build toolchain

| Tool              | Role                                                     |
| ----------------- | -------------------------------------------------------- |
| SvelteKit         | Framework + static site generation                       |
| Vite              | Build tool                                               |
| TypeScript        | Type safety                                              |
| TailwindCSS       | Utility CSS                                              |
| Flowbite-Svelte   | Component library                                        |
| ESLint + Prettier | Linting & formatting                                     |
| Playwright        | Browser tests + integration tests                        |
| Buf               | Generates TypeScript proto bindings from `server/proto/` |
