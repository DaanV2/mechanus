# scene-config-files Specification

## Purpose

Defines how authored scene and room configuration is stored on disk, versioned and migrated when the server starts.

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

### Requirement: Config files are read at start
The system SHALL read scene and room config files when the server starts and SHALL NOT watch them; changes to files on disk take effect after a restart.

#### Scenario: Edit a scene file
- **WHEN** the GM saves a change to a scene file while the server runs
- **THEN** the change takes effect after the server restarts

### Requirement: Invalid files are skipped and reported
When a config file is invalid, the system SHALL skip it, keep loading the other files, and SHALL report the file and error in the GM view.

#### Scenario: Syntax error
- **WHEN** a hand-edited scene file no longer parses
- **THEN** the server starts without that scene and the GM view shows the file and error
