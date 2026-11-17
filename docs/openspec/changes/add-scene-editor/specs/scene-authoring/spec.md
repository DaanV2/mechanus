# Spec Delta

## Purpose

Lets the GM create, edit and preview scenes, variants and the room setup in the app, saving to the authored configuration files.

## ADDED Requirements

### Requirement: Edit scenes in the app
The system SHALL let the GM create and edit a scene's entities, base instructions per target, and variants, saving them to the scene's config file.

#### Scenario: Add a variant
- **WHEN** the GM adds a `night` variant in the scene editor
- **THEN** the scene's config file contains the variant and it can be activated

### Requirement: Edit the room in the app
The system SHALL let the GM edit the room's devices and target bindings, saving them to the room config file.

#### Scenario: Add a light
- **WHEN** the GM adds a light device to the room
- **THEN** the room config file lists it and scenes can target it

### Requirement: Preview without activating
The system SHALL let the GM preview a scene or variant without activating it or instructing any device.

#### Scenario: Preview a variant
- **WHEN** the GM previews the `storm` variant
- **THEN** the GM sees what each target would show while room devices keep the active scene
