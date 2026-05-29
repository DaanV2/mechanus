# Contributing

## Tools

You'll need the following tools:

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Nodejs](https://nodejs.org/en)
- [Just](https://just.systems/man/en/packages.html) — command runner for project tasks

## Common commands

From the repo root:

```bash
just proto          # Generate Go code from proto files
just proto-format   # Format proto files
just proto-lint     # Lint proto files
just image          # Build the Docker image
just checks         # Run all checks (proto, format, lint, image)
```

From the `server/` directory:

```bash
just build          # Build all packages
just test           # Run all tests with coverage
just lint           # Lint and auto-fix code
just format         # Format Go source files
just generate       # Run code generation
just documentation  # Generate documentation
just start-server   # Start the server
just checks         # Run all checks (build, test, format, lint, docs)
```

Run `just --list` in any directory with a justfile to see all available commands with descriptions.
