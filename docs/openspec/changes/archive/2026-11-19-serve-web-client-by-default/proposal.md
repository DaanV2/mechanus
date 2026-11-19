# Proposal

## Why

Milestone M6 (Polish & release). Running the server without `--server.static.folder` serves no web client, which trips up new users.

## What Changes

- Serve the web client without the flag: embed it in the binary, or default to a `web/` folder next to the executable

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `distribution`: the web client is embedded in the binary and served without extra configuration; archives drop `web/`

## Impact

- server static file handling, release packaging, docs/getting-started/
