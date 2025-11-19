# Development

## Getting started

Install tools:

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Nodejs](https://nodejs.org/en)

if you working on frontend check [client/docs/development](../client/docs/development.md).  
if you working on backend check [server/docs/development](../server/docs/development.md).

## Docker Compose for Local Development

For local development, you can use Docker Compose with the watch feature to automatically rebuild the application when code changes are detected.

### Start with Watch Mode

To start the application with automatic rebuilds on code changes:

```bash
docker compose watch
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
docker compose down
```

### Ports

The following ports are exposed:
- `8080` - Main web server (HTTP)
- `8443` - Secure web server (HTTPS)
- `8666` - Additional service port
