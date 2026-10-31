# Development

## Setup

Requires [Go](https://go.dev/). From the repo root, `just --list` shows all available commands; `cd server` for the server-specific ones below.

## Common commands

```bash
just build           # go build ./...
just test            # Ginkgo unit tests, with coverage
just lint            # golangci-lint --fix
just format          # go fmt ./...
just generate        # go generate ./...
just start-server    # go run main.go server
just checks          # build + test + format + lint + documentation
```

## Running locally

```bash
go run main.go server
```

Config is loaded via Viper from flags, environment variables, or a config file — see [configuration/config-file.md](../../docs/configuration/config-file.md). The default SQLite database lives in the OS state directory; see [configuration/database.md](../../docs/configuration/database.md) to point it elsewhere or switch backends.

## Structure

See [architecture/server.md](../../docs/architecture/server.md) for the Clean Architecture layers (`engine`, `application`, `infrastructure`, `components`) and the startup/shutdown lifecycle.

## Proto changes

After editing a `.proto` file under `proto/`, run `just proto` from the repo root to regenerate Go and TypeScript stubs before building.
