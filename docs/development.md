# Development

## Getting started

Install tools:

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Nodejs](https://nodejs.org/en)
- [just](https://github.com/casey/just) — runs the recipes in the `.justfile`s (`just --list`)
- [OpenSpec](https://github.com/Fission-AI/OpenSpec) — specs and planned changes in [docs/openspec](./openspec/README.md):
  `npm install -g @fission-ai/openspec@2.1.0` (keep in step with `OPENSPEC_VERSION` in
  `.github/workflows/openspec-checks.yaml`), then `just spec list`

if you working on frontend check [client/docs/development](../client/docs/development.md).  
if you working on backend check [server/docs/development](../server/docs/development.md).

## Docker Compose for Local Development

For local development, you can use Docker Compose with the watch feature to automatically rebuild the application when code changes are detected.

### Start with Watch Mode

To start the application with automatic rebuilds on code changes:

```bash
docker compose watch # OR just dev
```

This will:

- Build and start the mechanus container
- Watch for changes in:
  - Server Go files (`./server`)
  - Client source files (`./client/src`)
  - Client configuration files (package.json, vite.config.ts, etc.)
  - Shared proto files (`./shared`)
- Automatically rebuild the container when relevant files change

### Regular Start (without watch)

To start without watch mode:

```bash
docker compose up
```

### Stop and Clean Up

```bash
docker compose down # OR just dev-stop
```

### Ports

The following ports are exposed:

- `8080` - Main web server (HTTP). Configurable via `--server.port` / `SERVER_PORT`.
- `8443` - Reserved in `docker-compose.yaml` for a future HTTPS listener; the server does not yet bind to it.
