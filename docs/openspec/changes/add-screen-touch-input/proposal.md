# Proposal

## Why

Milestone M3 (Live play on the battle map). Table screens are often touch screens; players should be able to interact with them within limits the GM controls.

## What Changes

- Per-device `touch` room setting: `none`, `view` (pan, zoom, ping, measure; default) or `control` (move tokens as a configured role or user)
- GM can lock touch input live

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-maps`: adds touch input modes on screens

## Impact

- room config, screen WebSocket validation, client input handling
