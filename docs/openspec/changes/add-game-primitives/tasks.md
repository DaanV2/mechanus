# Tasks

## 1. Primitives

- [ ] 1.1 Add bar definitions to campaign settings and bar values to scene state; verify with tests
- [ ] 1.2 Add condition tags with round counters; verify counter expiry with tests
- [ ] 1.3 Add initiative order with turn / round tracking; verify with tests
- [ ] 1.4 Add free-form stats; verify with tests
- [ ] 1.5 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 1.6 Run `just proto-lint` after proto changes and fix every reported issue

## 2. Overlays and GM view

- [ ] 2.1 Render the `initiative` overlay; verify with a Playwright test
- [ ] 2.2 Add portraits and the `entities` overlay on all views; verify visually and with a Playwright test
- [ ] 2.3 Add GM controls for bars, conditions, initiative and stats; verify with a Playwright test
- [ ] 2.4 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
