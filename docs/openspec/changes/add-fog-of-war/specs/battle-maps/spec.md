# Spec Delta

## ADDED Requirements

### Requirement: Fog of war
The system SHALL cover unrevealed areas of a battle map with fog on player-facing screens and SHALL NOT send entities inside fog to them.

#### Scenario: Goblin in fog
- **WHEN** an entity stands in an unrevealed area
- **THEN** player-facing screens neither show nor receive it

### Requirement: Reveal and hide areas
The system SHALL let the GM reveal and re-hide areas of the map, storing the fog in scene state.

#### Scenario: Reveal a room
- **WHEN** the GM reveals a room
- **THEN** it becomes visible on every player-facing screen and stays revealed after the scene is re-activated
