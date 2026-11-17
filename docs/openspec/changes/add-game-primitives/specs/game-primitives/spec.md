# Spec Delta

## Purpose

Gives the GM generic, system-agnostic tools to track entity resources, conditions, turn order and stats during play.

## ADDED Requirements

### Requirement: Resource bars
The system SHALL let a campaign define resource bars (current / max) and show each entity's bars to everyone, its owner only, or the GM only, as configured per bar.

#### Scenario: Owner-only HP
- **WHEN** the `HP` bar is owner-only
- **THEN** only the controlling player and the GM see a character's HP

### Requirement: Condition tags
The system SHALL let users define condition tags with an icon and an optional round counter, and attach them to entities.

#### Scenario: Counter runs out
- **WHEN** a condition with a 2-round counter is attached and two rounds pass
- **THEN** the condition is removed

### Requirement: Initiative tracking
The system SHALL keep an initiative order from hand-entered values and track the current turn and round.

#### Scenario: Next turn
- **WHEN** the GM advances the turn on the last entity in the order
- **THEN** the round number increases and the first entity is current

### Requirement: Entity stats
The system SHALL let the GM attach free-form key / value stats to entities.

#### Scenario: Add a stat
- **WHEN** the GM adds `AC: 15` to a goblin
- **THEN** the stat is stored with the entity in scene state
