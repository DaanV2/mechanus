# Tasks

## 1. Server

- [ ] 1.1 Store `controlled_by` and the movement lock in scene state; verify with state tests
- [ ] 1.2 Validate token moves against control and lock; verify with screen WebSocket tests
- [ ] 1.3 Add GM RPCs to assign control and toggle the lock; verify with tests
- [ ] 1.4 Run `just server lint` and `just server test` and fix every reported issue
- [ ] 1.5 Run `just proto-lint` after proto changes and fix every reported issue

## 2. Client

- [ ] 2.1 Add control assignment and lock toggle to the GM view; verify with a Playwright test
- [ ] 2.2 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
