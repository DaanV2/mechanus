# Tasks

## 1. API

- [ ] 1.1 Add authoring RPCs to `scenes/v1` and generate stubs with `just proto`; verify the build
- [ ] 1.2 Implement save-to-file in the application layer; verify with tests that round-trip a scene
- [ ] 1.3 Run `just proto-lint` after proto changes and fix every reported issue

## 2. Client

- [ ] 2.1 Build the scene editor (entities, base instructions, variants); verify with a Playwright test
- [ ] 2.2 Build the room editor (devices, bindings); verify with a Playwright test
- [ ] 2.3 Add scene / variant preview; verify no device receives instructions during preview
- [ ] 2.4 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 2.5 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in docs/roadmap.md.
