# Proposal

## Why

Milestone M4 (The room comes alive). Scenes can describe lighting but nothing drives real lights yet. Lights: Matter is one of three separate native light integrations (Hue, Home Assistant, Matter), built on the existing driver framework.

## What Changes

- Driver that applies scene light instructions to Matter lights through a native Go controller (commissioning, fabric management)
- Integration config set `integrations.matter` with fabric credentials
- Device status reported to the GM

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `device-drivers`: adds the Matter light integration

## Impact

- `server/infrastructure/outputs/matter`
- `components/outputs.go`
- docs/configuration/integrations.md
