# Tasks

## 1. Home Assistant

- [ ] 1.1 Add the `integrations.home-assistant` config set and register it in `cmd/server.go`; verify config docs are generated
- [ ] 1.2 Implement the Home Assistant output driver and register it with `DeviceMonitor.Track`; verify with tests against a fake
- [ ] 1.3 Document setup in docs/configuration/integrations.md
- [ ] 1.4 Run `just server lint` and `just server test` and fix every reported issue

## Workflow follow-up

- Archive the change and update its entry in the milestone list in docs/openspec/README.md.
