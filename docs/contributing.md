# Contributing

## Tools

You'll need the following tools:

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Nodejs](https://nodejs.org/en)
- [Just](https://just.systems/man/en/packages.html) — command runner for project tasks

## Common commands

The root `justfile` imports `server/justfile` and `client/justfile` as modules, so every
recipe is runnable from the repo root — no need to `cd` into a subdirectory.

Root-level recipes:

```bash
just proto          # Generate Go code from proto files
just proto-format   # Format proto files
just proto-lint     # Lint proto files
just image          # Build the Docker image
just checks         # Run all checks (proto, format, lint, image)
```

Server recipes (run from the repo root as `just server <recipe>`):

```bash
just server build          # Build all packages
just server test           # Run all tests with coverage
just server lint           # Lint and auto-fix code
just server format         # Format Go source files
just server generate       # Run code generation
just server documentation  # Generate documentation
just server start-server   # Start the server
just server checks         # Run all checks (build, test, format, lint, docs)
```

Client recipes (run from the repo root as `just client <recipe>`):

```bash
just client build   # Build the client
just client dev     # Start the Vite dev server
just client lint     # Lint client source files
just client checks  # Format and lint the client
```

`just server <recipe>` and `just client <recipe>` run as if you had `cd`'d into that
directory first — `just server::test` (double-colon) also works if you prefer that form.

Run `just --list` to see all top-level recipes and modules, or `just --list server` /
`just --list client` to see a module's recipes.
