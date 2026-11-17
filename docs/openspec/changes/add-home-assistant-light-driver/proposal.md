# Proposal

## Why

Milestone M4 (The room comes alive). Scenes can describe lighting but nothing drives real lights yet. Lights: Home Assistant is one of three separate native light integrations (Hue, Home Assistant, Matter), built on the existing driver framework.

## What Changes

- Driver that applies scene light instructions to lights exposed by Home Assistant
- Integration config set `integrations.home-assistant` with a Home Assistant URL and access token
- Device status reported to the GM

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `device-drivers`: adds the Home Assistant light integration

## Impact

- `server/infrastructure/outputs/home-assistant`
- `components/outputs.go`
- docs/configuration/integrations.md
