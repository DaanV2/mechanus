# Proposal

## Why

Milestone M3 (Live play on the battle map). Running combat needs hit points, conditions and turn order on screen. Mechanus has no rules engine and no dice, so it offers generic primitives instead.

## What Changes

- Campaign-defined resource bars (current / max) with visibility everyone / owner / GM
- User-defined condition tags with icons and an optional round counter
- Initiative order with hand-entered values, turn and round tracking
- Free-form key / value entity stats
- Overlays: the `initiative` tracker, portraits in the `entities` overlay, and the `entities` overlay on views other than the battle map

## Capabilities

### New Capabilities

- `game-primitives`: resource bars, conditions, initiative and stats on entities

### Modified Capabilities

- `screen-views`: adds the initiative overlay, entity portraits, and the entities overlay on all views

## Impact

- scene state, campaign settings, `scenes/v1`, overlays in the client renderer, GM view
