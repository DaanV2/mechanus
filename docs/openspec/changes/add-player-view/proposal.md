# Proposal

## Why

Milestone M3 (Live play on the battle map). Players need their own screen showing only what their character may see.

## What Changes

- Player view derived server-side per user from battle map state
- Hidden and fogged entities are never sent; camera follows their character
- A scene's `players` target can override the derived view for special moments

## Capabilities

### New Capabilities

- `player-view`: the per-player screen and its visibility rules

### Modified Capabilities

None.

## Impact

- screen driver, client player route, scene targets
