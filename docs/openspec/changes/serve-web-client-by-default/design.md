# Design

## Context

Release archives already ship `web/` next to the binary, but the server only serves it when started with `--server.static.folder web`; the default `/web` only exists in the Docker image.

## Decisions

- **Embed the client and install it on disk.** The compiled client is embedded in the binary (`server/infrastructure/webclient`, `//go:embed all:dist`), so the binary alone is enough to run Mechanus. On start the server makes sure the static folder holds the client, in the spirit of `InitializeServer`:
  - folder missing → create it and write the embedded files into it, plus a `.mechanus-client` marker holding a hash of the embedded files;
  - folder exists with a marker of another hash (the binary was updated) → replace it with the embedded files;
  - folder exists without the marker → it is the user's (Docker image, development, a custom build): serve it untouched;
  - the folder can't be written → serve the embedded files straight from memory and log a warning.
- **Default folder `web`**, relative to the working directory (the folder the server starts from, like `.config` / `.local`), instead of `/web`. `--server.static.folder` still overrides it; the Docker image keeps `SERVER_STATIC_FOLDER=/web`.
- **Builds without the client still compile.** `dist/` holds only a committed `.gitkeep`; the rest is git-ignored. `go run ./scripts/embedclient` copies `client/build` into it before a release build (GoReleaser hook, `just binary`). A binary built without it has nothing embedded and serves the static folder as before.
- Release archives no longer ship `web/`: the binary is the whole program.

## Open Questions

None.
