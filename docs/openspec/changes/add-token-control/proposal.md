# Proposal

## Why

Milestone M3 (Live play on the battle map). Any logged-in user can move any token today. Players should move only what they control, and the GM needs to share control and freeze movement.

## What Changes

- Each entity gets a `controlled_by` list of users (a character defaults to its player)
- GM assigns and shares control
- GM can lock all player movement live
- Token moves are checked against control server-side

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-maps`: token moves are limited to controllers; adds GM control assignment and a movement lock

## Impact

- scene state, `scenes/v1`, screen WebSocket validation, GM view
