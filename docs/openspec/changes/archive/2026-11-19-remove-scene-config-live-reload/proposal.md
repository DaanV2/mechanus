# Proposal

## Why

Milestone M2 (Scene authoring). Live reload of scene and room files (from `add-versioned-config-files`) watched the config folders with fsnotify and re-instructed devices when the active scene or room changed. It costs more than it gives: watching is unreliable on bind mounts with Docker Desktop on Windows / macOS, re-applying a half-edited active scene mid-session changes the table under the players, and the scene editor (`add-scene-editor`) will save through the server, so the server already knows when configuration changes.

Scene and room files are user files: GMs edit, create and import them while the server runs. Those changes should show up without a restart, so the files are read from disk when they are needed instead: listing scenes scans the folder, activating or resetting a scene reads its file, and the room file is checked whenever the room is used.

## What Changes

- Remove the config folder watcher, `ConfigStore.Reload` and its change listeners
- Remove `SceneService.SceneConfigChanged` / `RoomConfigChanged`; devices are no longer re-instructed when a file changes, the active scene keeps the configuration it was activated with
- Read scene files on demand: `ListSceneConfigs` scans the scenes folder on every call, `GetSceneConfig` (activate, reset) reads the scene's file
- Read the room file on demand in `CurrentRoom`, parsing it again only when its name, size or modification time changed
- An invalid scene or room file is skipped and reported in the GM view the next time scenes are listed, instead of keeping its last good version
- Remove the Playwright test that edited files in the running container; mount the test scenes folder read-only again

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scene-config-files`: live reload removed; scene and room files are read on demand, and invalid files are skipped

## Impact

- `server/infrastructure/scenestore`, `server/application/scene_service.go`, `server/components`
- `scenes/v1` proto comment on `config_errors`, GM view error text
- `tests/` compose files and config file tests
- docs/configuration/scene-files.md, docs/architecture/scenes.md
