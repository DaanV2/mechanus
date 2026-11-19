# Tasks

## 1. Server

- [x] 1.1 Remove the fsnotify watcher, `ConfigStore.Reload` and the change listeners; `Load` reads every file fresh and skips invalid ones; verify with component tests for skipped, fixed and deleted files
- [x] 1.2 Remove `SceneService.SceneConfigChanged` / `RoomConfigChanged` and their wiring in `components`
- [x] 1.3 Run `just server lint`, `just server test`, `just server vet` and `just server vulncheck` and fix every reported issue

## 2. Client and tests

- [x] 2.1 Update the GM view error text and the `config_errors` proto comment; drop the scene list refresh timer
- [x] 2.2 Remove the Playwright config file edit test and make the test scenes mount read-only
- [x] 2.3 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## 3. Docs

- [x] 3.1 Update docs/configuration/scene-files.md and docs/architecture/scenes.md

## Workflow follow-up

- Archive the change.
