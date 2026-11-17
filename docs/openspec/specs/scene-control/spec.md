# scene-control Specification

## Purpose

Lets the GM run the room through scenes: pick the active scene, switch its variant, and have every bound device follow, while play changes are kept as scene state separate from the authored configuration.

## Requirements

### Requirement: One room per instance
The system SHALL serve exactly one room per running instance, read from the room file; a missing room file SHALL be treated as an empty room.

#### Scenario: No room file
- **WHEN** the server starts without a room file
- **THEN** it runs with an empty room and no device receives instructions

### Requirement: List scenes
The system SHALL let any logged-in user list the available scenes with their variants, including in the GM view at `/views/game-master/`.

#### Scenario: GM opens the GM view
- **WHEN** a logged-in user opens the GM view
- **THEN** the available scenes and their variants are listed

### Requirement: Activate a scene
The system SHALL let any logged-in user activate a scene, which saves the state of the previously active scene and instructs every bound device with the new scene's effective instructions.

#### Scenario: Activate
- **WHEN** the GM activates the `cave` scene
- **THEN** each device bound to a target of `cave` receives its instruction for the scene's current variant

### Requirement: Switch variant
The system SHALL let any logged-in user switch the active scene's variant, re-instructing only the targets whose effective instruction changed.

#### Scenario: Night falls
- **WHEN** the GM switches the active scene from `day` to `night`
- **THEN** devices whose instruction differs between the variants are updated and the others are left alone

### Requirement: Get active scene state
The system SHALL return the active scene with its variant and entities on request.

#### Scenario: GM view reloads
- **WHEN** the GM view requests the active scene
- **THEN** it receives the active scene, its variant and its entities

### Requirement: Configuration is read-only at runtime
The system SHALL NOT modify authored scene or room configuration during play; every play change SHALL be stored as scene state holding only the difference from the configuration.

#### Scenario: Entity moved
- **WHEN** a token is moved during play
- **THEN** the scene's configuration files are unchanged and the move is recorded in scene state

### Requirement: Scene state is persisted on load and exit
The system SHALL load a scene's saved state when it is activated and save it when the scene is exited, including when the server shuts down.

#### Scenario: Shutdown with an active scene
- **WHEN** the server shuts down while a scene is active
- **THEN** the scene's state is saved and restored the next time that scene is activated

### Requirement: Reset a scene
The system SHALL let the GM reset a scene's state back to its configuration, discarding spawned entities, moves, visibility changes, overrides, cameras and the chosen variant.

#### Scenario: Reset button
- **WHEN** the GM presses reset on a scene in the GM view
- **THEN** that scene returns to its authored configuration and any bound devices are re-instructed if it is active
