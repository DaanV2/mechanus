# Tasks

## 1. Formats and versioning

- [ ] 1.1 Add YAML / JSON loading by extension in the scene store; verify with tests for both formats
- [ ] 1.2 Add the `version` field and a migration chain; verify with tests for older, current and newer versions
- [ ] 1.3 Document the file format in docs/configuration/ and verify the examples load

## 2. Live reload

- [ ] 2.1 Watch scene / room folders and reload changed files; verify with a test that edits a file and observes the reload
- [ ] 2.2 Re-instruct devices when the active scene changed; verify with a component test
- [ ] 2.3 Report invalid files in the GM view while keeping the last good version; verify with a Playwright test
- [ ] 2.4 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 2.5 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
