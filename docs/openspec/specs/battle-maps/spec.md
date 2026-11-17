# battle-maps Specification

## Purpose

Defines the battle map: an authored file of grid settings and typed layers that the server reads and turns into what each screen may see, plus the live interaction on it (tokens, camera, entity visibility).

## Requirements

### Requirement: Battle map file format v0
The system SHALL read battle maps from `*.map.json` files in a scene's folder, consisting of grid settings and an ordered list of image and entity layers, where an instance group places one asset at many positions.

#### Scenario: Map with an instance group
- **WHEN** a map declares an instance group of 40 trees
- **THEN** the screen draws all 40 from one asset

### Requirement: Map selected by instruction asset
The system SHALL pick a screen's battle map from the `asset` of its screen instruction, so variants can swap maps.

#### Scenario: Variant swaps the map
- **WHEN** the `flooded` variant names `maps/cave-flooded.map.json`
- **THEN** switching to it shows the flooded map

### Requirement: Asset paths
The system SHALL resolve asset paths relative to the scene folder (`maps/cave.webp`) or from a shared folder when prefixed (`shared:/monsters/goblin.png`).

#### Scenario: Shared asset
- **WHEN** a map references `shared:/monsters/goblin.png`
- **THEN** the screen loads it from the shared assets URL

### Requirement: Map files are not served directly
The system SHALL NOT serve battle map files under `/assets/`; screens SHALL receive only the layers and entities they may see.

#### Scenario: Hidden entity
- **WHEN** a map has an entity that is hidden in the scene state
- **THEN** no screen receives that entity

### Requirement: Entity tokens
The system SHALL draw entity tokens at the position from the map file, unless the scene state holds a moved position, which takes precedence.

#### Scenario: Token moved earlier
- **WHEN** a scene is re-activated after a token was moved
- **THEN** the token is drawn at its moved position

### Requirement: Token moves, pan and zoom from screens
The system SHALL accept token moves, pans and zooms from screens over the screen WebSocket, validate them server-side, and send the result to the screens of the scene without reloading the map. Paired devices SHALL only pan and zoom.

#### Scenario: User drags a token
- **WHEN** a logged-in user moves a token on a screen
- **THEN** the move is validated, stored in scene state and shown on every screen of the scene

#### Scenario: Device tries to move a token
- **WHEN** a paired device sends a token move
- **THEN** the move is rejected

### Requirement: Camera per target
The system SHALL keep the camera (pan and zoom) per target in the scene state.

#### Scenario: Camera restored
- **WHEN** a scene is re-activated
- **THEN** each target's screen opens with its last camera

### Requirement: GM entity controls
The system SHALL let the GM reveal, hide, spawn and remove entities through the scene API and the GM view, storing the change in scene state.

#### Scenario: Reveal an ambush
- **WHEN** the GM reveals a hidden goblin
- **THEN** the goblin appears on every screen of the scene
