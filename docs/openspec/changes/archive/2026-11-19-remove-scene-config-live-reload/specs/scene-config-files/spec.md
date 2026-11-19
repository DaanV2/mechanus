# Spec Delta

## REMOVED Requirements

### Requirement: Live reload
**Reason**: Watching is unreliable on Docker Desktop bind mounts, and re-applying an edited active scene mid-session disrupts play. The scene editor will save through the server instead.
**Migration**: Edited scene and room files are picked up the next time they are used, see "Scene and room files are read on demand".

### Requirement: Invalid files keep the last good version
**Reason**: Files are read when needed, so there is no earlier version to fall back to.
**Migration**: See "Invalid files are skipped and reported".

## ADDED Requirements

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
