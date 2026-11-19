# Proposal

## Why

Milestone M2 (Scene authoring). Scene and room files must be hand-written today, in a single format with no versioning, and edits need a restart. Authoring (editor or by hand) needs a stable, evolvable file format and live reload.

## What Changes

- Accept scene / room config files in YAML or JSON, picked by extension
- Add a `version` field and migrate older files on load
- Translate config to the `scenes/v1` proto for the API and screens; the proto stays the wire contract, not the file schema
- Watch config folders, reload changed files, re-instruct devices when the active scene changed
- Report invalid files in the GM view and keep the last good version

## Capabilities

### New Capabilities

- `scene-config-files`: file formats, versioning, migrations and live reload of authored scene / room configuration

### Modified Capabilities

None.

## Impact

- `server/infrastructure/persistence/scenestore`, file watching, `scenes/v1` proto
- GM view error reporting
- docs/configuration/
