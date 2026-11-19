# Tasks

## 1. Server

- [x] 1.1 Remove the fsnotify watcher, `ConfigStore.Reload` and the change listeners
- [x] 1.2 Read scene files on demand in `ListSceneConfigs` and `GetSceneConfig`, and the room file in `CurrentRoom` (parsed again only when it changed); skip and report invalid files; verify with component tests for added, edited, deleted, invalid and fixed scene and room files, and for activating a scene after editing it or the room
- [x] 1.3 Remove `SceneService.SceneConfigChanged` / `RoomConfigChanged` and their wiring in `components`
- [x] 1.4 Run `just server lint`, `just server test`, `just server vet` and `just server vulncheck` and fix every reported issue

## 2. Client and tests

- [x] 2.1 Update the GM view error text and the `config_errors` proto comment; drop the scene list refresh timer
- [x] 2.2 Remove the Playwright config file edit test and make the test scenes mount read-only
- [x] 2.3 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## 3. Docs

- [x] 3.1 Update docs/configuration/scene-files.md and docs/architecture/scenes.md

## Workflow follow-up

- Archive the change.
