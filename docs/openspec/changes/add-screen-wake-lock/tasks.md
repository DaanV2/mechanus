# Tasks

## 1. Wake lock

- [ ] 1.1 Acquire a Screen Wake Lock in the device view when in a secure context and re-acquire on `visibilitychange`; verify on an HTTPS setup
- [ ] 1.2 Add the plain-HTTP fallback; verify a TV over plain HTTP stays awake past its sleep timeout
- [ ] 1.3 Run `just client lint` and `npm run check` in client/ and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
