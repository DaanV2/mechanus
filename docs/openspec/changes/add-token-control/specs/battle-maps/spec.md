# Spec Delta

## ADDED Requirements

### Requirement: Token control
The system SHALL only accept a token move from a user listed in the entity's `controlled_by` or from the GM; a character SHALL default to being controlled by its player.

#### Scenario: Player moves own token
- **WHEN** a player moves their character's token
- **THEN** the move is accepted

#### Scenario: Player moves a goblin
- **WHEN** a player moves an entity they do not control
- **THEN** the move is rejected

### Requirement: GM assigns control
The system SHALL let the GM add or remove users from an entity's `controlled_by`.

#### Scenario: Share a summon
- **WHEN** the GM adds two players to a summoned creature
- **THEN** both can move it

### Requirement: Movement lock
The system SHALL let the GM freeze and unfreeze all player token movement during play.

#### Scenario: Freeze
- **WHEN** the GM locks movement
- **THEN** player token moves are rejected until unlocked, while the GM can still move tokens
