# Development

## Setup

```bash
cd client
npm install
```

## Common commands

```bash
npm run dev     # vite dev server
npm run build   # production build -> served by the Go server
npm run lint     # ESLint
npm run check    # svelte-check + tsc
```

## Proto bindings

TypeScript proto stubs under `src/proto/` are generated from `server/proto/` via Buf. Run `just proto` from the repo root after changing any `.proto` file, then restart the dev server.

## Structure

See [architecture/client.md](../../docs/architecture/client.md) for view types, rendering, and the server communication layers (`lib/api/`, `lib/networking/`).

## Running against the real server

`npm run dev` only serves the SvelteKit app; it does not proxy API calls. Start a Mechanus server separately (`cd server && go run main.go server`, or `docker compose watch` from the repo root) and point the client at it via `lib/config` (see `server_grpc_url`). See [docs/development.md](../../docs/development.md) for the full local dev guide.
