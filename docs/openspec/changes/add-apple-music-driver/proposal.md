# Proposal

## Why

Milestone M4 (The room comes alive). Scenes should start the right music. iTunes / Apple Music is one of the planned music integrations, built on the existing driver framework.

## What Changes

- Driver that plays, pauses and sets volume on iTunes / Apple Music from scene music instructions
- Integration config set `integrations.apple-music`
- Account linking flow for the GM

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `device-drivers`: adds the iTunes / Apple Music music integration

## Impact

- `server/infrastructure/outputs/apple-music`
- GM account linking UI
- docs/configuration/integrations.md
