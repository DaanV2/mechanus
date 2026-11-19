# Proposal

## Why

Milestone M2 (Scene authoring). Live reload of scene and room files (from `add-versioned-config-files`) watched the config folders with fsnotify and re-instructed devices when the active scene or room changed. It costs more than it gives: watching is unreliable on bind mounts with Docker Desktop on Windows / macOS, re-applying a half-edited active scene mid-session moves the table under the players, and the scene editor (`add-scene-editor`) will save through the server, so the server already knows when configuration changes. Reading files once at start keeps the config store simple until then.

## What Changes

- Remove the config folder watcher, `ConfigStore.Reload` and its change listeners
- Remove `SceneService.SceneConfigChanged` / `RoomConfigChanged`; devices are no longer re-instructed when a file changes
- Read scene and room files once when the server starts; edits need a restart
- An invalid file is skipped and reported in the GM view, instead of keeping its last good version (there is none without reloading)
- Remove the Playwright test that edited files in the running container; mount the test scenes folder read-only again

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scene-config-files`: live reload removed; files are read at start and invalid files are skipped

## Impact

- `server/infrastructure/scenestore`, `server/application/scene_service.go`, `server/components`
- `scenes/v1` proto comment on `config_errors`, GM view error text
- `tests/` compose files and config file tests
- docs/configuration/scene-files.md, docs/architecture/scenes.md
