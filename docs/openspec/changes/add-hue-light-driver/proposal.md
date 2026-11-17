# Proposal

## Why

Milestone M4 (The room comes alive). Scenes can describe lighting but nothing drives real lights yet. Lights: Philips Hue is one of three separate native light integrations (Hue, Home Assistant, Matter), built on the existing driver framework.

## What Changes

- Driver that applies scene light instructions to lights through a Philips Hue bridge
- Integration config set `integrations.hue` with a Hue bridge address and app key
- Device status reported to the GM

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `device-drivers`: adds the Philips Hue light integration

## Impact

- `server/infrastructure/outputs/hue`
- `components/outputs.go`
- docs/configuration/integrations.md
