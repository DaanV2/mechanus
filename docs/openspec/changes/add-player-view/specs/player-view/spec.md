# Spec Delta

## Purpose

Gives each player a personal screen that shows the battle map from their character's point of view, limited to what they may see.

## ADDED Requirements

### Requirement: Derived player view
The system SHALL derive each player's view server-side from the battle map state, never sending hidden or fogged entities, with the camera on their character.

#### Scenario: Player opens their view
- **WHEN** a player opens the player view during a battle map scene
- **THEN** they see the map centred on their character without hidden or fogged entities

### Requirement: Scene override
When a scene instructs the `players` target, the system SHALL show that instruction on player views instead of the derived view.

#### Scenario: Cutscene
- **WHEN** the scene sends a scenery video to `players`
- **THEN** every player view plays the video until the instruction is removed
