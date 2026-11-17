# Proposal

## Why

Milestone M5 (Sessions & campaigns). Scene entities representing player characters should be tied to the campaign's characters so control, player views and stats follow the right player.

## What Changes

- Characters belong to a campaign and a player
- Scene entities can reference a character (`Entity.CharacterID`)
- A character entity defaults to being controlled by its player

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `campaigns`: adds characters owned by players

## Impact

- campaign models, scene entities, GM view
