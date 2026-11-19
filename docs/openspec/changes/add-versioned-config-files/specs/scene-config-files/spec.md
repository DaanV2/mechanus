# Spec Delta

## Purpose

Defines how authored scene and room configuration is stored on disk, versioned, migrated, and reloaded while the server runs.

## ADDED Requirements

### Requirement: YAML and JSON config files
The system SHALL read scene and room configuration from YAML or JSON files, choosing the parser by file extension.

#### Scenario: YAML scene
- **WHEN** a scene folder contains `scene.yaml`
- **THEN** the scene is loaded the same as its JSON equivalent

### Requirement: Versioned config with migrations
Every config file SHALL carry a `version` field; the system SHALL migrate files with an older version on load without rewriting them, and SHALL reject files with an unknown newer version.

#### Scenario: Older file
- **WHEN** a scene file has an older `version`
- **THEN** it is migrated in memory and loads normally

#### Scenario: Newer file
- **WHEN** a scene file has a `version` newer than the server supports
- **THEN** the file is reported as invalid

### Requirement: Live reload
The system SHALL watch config folders, reload changed files, and re-instruct devices when the active scene's configuration changed.

#### Scenario: Edit the active scene
- **WHEN** the GM saves a change to the active scene's file
- **THEN** bound devices receive the updated instructions without a restart

### Requirement: Invalid files keep the last good version
When a changed file is invalid, the system SHALL keep using its last good version and SHALL report the error in the GM view.

#### Scenario: Syntax error
- **WHEN** a hand-edited scene file no longer parses
- **THEN** the previous version stays active and the GM view shows the file and error
