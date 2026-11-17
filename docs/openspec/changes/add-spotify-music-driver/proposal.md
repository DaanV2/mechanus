# Proposal

## Why

Milestone M4 (The room comes alive). Scenes should start the right music. Spotify is one of the planned music integrations, built on the existing driver framework.

## What Changes

- Driver that plays, pauses and sets volume on Spotify from scene music instructions
- Integration config set `integrations.spotify`
- Account linking flow for the GM

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `device-drivers`: adds the Spotify music integration

## Impact

- `server/infrastructure/outputs/spotify`
- GM account linking UI
- docs/configuration/integrations.md
