# Tasks

## 1. State keying

- [ ] 1.1 Key scene state by campaign + scene with a migration for existing state; verify with migration tests
- [ ] 1.2 Run `just server lint` and `just server test` and fix every reported issue

## 2. Sessions

- [ ] 2.1 Add the session object and campaign switching; verify with tests
- [ ] 2.2 Persist and restore the active scene on restart; verify with a component test
- [ ] 2.3 Add a campaign switcher to the GM view; verify with a Playwright test
- [ ] 2.4 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in docs/roadmap.md.
