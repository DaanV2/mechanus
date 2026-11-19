# Proposal

## Why

Milestone M2 (Scene authoring). Scene and room files must be hand-written today, in a single format with no versioning. Authoring (editor or by hand) needs a stable, evolvable file format.

## What Changes

- Accept scene / room config files in YAML or JSON, picked by extension
- Add a `version` field and migrate older files on load
- Translate config to the `scenes/v1` proto for the API and screens; the proto stays the wire contract, not the file schema
- Report invalid files in the GM view and skip them

Config files are read when the server starts; there is no live reload, so edits to files on disk need a restart.

## Capabilities

### New Capabilities

- `scene-config-files`: file formats, versioning and migrations of authored scene / room configuration

### Modified Capabilities

None.

## Impact

- `server/infrastructure/persistence/scenestore`, `scenes/v1` proto
- GM view error reporting
- docs/configuration/
