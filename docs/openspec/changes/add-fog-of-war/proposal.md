# Proposal

## Why

Milestone M3 (Live play on the battle map). Encounters need unexplored areas hidden from players until revealed.

## What Changes

- Fog layer on battle maps
- GM reveals and hides areas
- Fog state stored in scene state
- Fogged areas and entities in them are never sent to player-facing screens

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-maps`: adds fog of war and reveal

## Impact

- battle map layers, scene state, screen driver filtering, client renderer
