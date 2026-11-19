# scene-config-files Specification

## Purpose

Defines how authored scene and room configuration is stored on disk, versioned and migrated when it is read.

## Requirements

### Requirement: YAML and JSON config files
The system SHALL read scene and room configuration from YAML or JSON files, choosing the parser by file extension.

#### Scenario: YAML scene
- **WHEN** the scenes folder contains `tavern.yaml`
- **THEN** scene `tavern` is loaded the same as its JSON equivalent `tavern.json`

#### Scenario: Legacy file
- **WHEN** the scenes folder contains a `.dat` file from an older version
- **THEN** the file is reported as invalid, with the advice to rename it to `.json` and add a version

### Requirement: Versioned config with migrations
Every config file SHALL carry a `version` field; the system SHALL migrate files with an older version on load without rewriting them, and SHALL reject files with an unknown newer version.

#### Scenario: Older file
- **WHEN** a scene file has an older `version`
- **THEN** it is migrated in memory and loads normally

#### Scenario: Newer file
- **WHEN** a scene file has a `version` newer than the server supports
- **THEN** the file is reported as invalid

### Requirement: Scene and room files are read on demand
Scene and room files are user files that can be edited, created or imported while the server runs. The system SHALL read them from disk when they are needed and SHALL NOT watch them: listing scenes SHALL scan the scenes folder, activating or resetting a scene SHALL read that scene's file, and using the room SHALL read the room file as it is then. The active scene SHALL keep the configuration it was activated with.

#### Scenario: Add a scene file
- **WHEN** the GM adds a scene file while the server runs
- **THEN** the scene is listed the next time scenes are listed

#### Scenario: Edit a scene file
- **WHEN** the GM edits the file of a scene and then activates the scene
- **THEN** the scene is activated with the edited configuration

#### Scenario: Edit the active scene's file
- **WHEN** the GM edits the file of the active scene
- **THEN** devices are not re-instructed until the scene is activated or reset again

#### Scenario: Edit the room file
- **WHEN** the GM binds another device in the room file and then activates a scene
- **THEN** the newly bound device is instructed too

### Requirement: Invalid files are skipped and reported
When a config file is invalid, the system SHALL skip it, keep loading the other files, and SHALL report the file and error in the GM view.

#### Scenario: Syntax error
- **WHEN** a hand-edited scene file no longer parses
- **THEN** the scene is not listed and the GM view shows the file and error
