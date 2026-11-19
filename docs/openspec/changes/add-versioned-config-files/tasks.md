# Tasks

## 1. Formats and versioning

- [ ] 1.1 Add YAML / JSON loading by extension in the scene store; verify with tests for both formats
- [ ] 1.2 Add the `version` field and a migration chain; verify with tests for older, current and newer versions
- [ ] 1.3 Document the file format in docs/configuration/ and verify the examples load

## 2. Invalid files

- [ ] 2.1 Skip invalid files on load and report them in the GM view; verify with a Playwright test
- [ ] 2.2 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 2.3 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
