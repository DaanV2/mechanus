# Spec Delta

## REMOVED Requirements

### Requirement: Live reload
**Reason**: Watching is unreliable on Docker Desktop bind mounts, and re-applying an edited active scene mid-session disrupts play. The scene editor will save through the server instead.
**Migration**: Edited scene files are picked up the next time scenes are listed or a scene is activated, see "Scene files are read on demand". Restart the server after editing the room file.

### Requirement: Invalid files keep the last good version
**Reason**: Scene files are read when needed, so there is no earlier version to fall back to.
**Migration**: See "Invalid files are skipped and reported".

## ADDED Requirements

### Requirement: Scene files are read on demand
The system SHALL read scene files from disk when they are needed and SHALL NOT watch them: listing scenes SHALL scan the scenes folder, and activating or resetting a scene SHALL read that scene's file. The active scene SHALL keep the configuration it was activated with. The room file SHALL be read when the server starts.

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
- **WHEN** the GM edits the room file while the server runs
- **THEN** the change takes effect after the server restarts

### Requirement: Invalid files are skipped and reported
When a config file is invalid, the system SHALL skip it, keep loading the other files, and SHALL report the file and error in the GM view.

#### Scenario: Syntax error
- **WHEN** a hand-edited scene file no longer parses
- **THEN** the scene is not listed and the GM view shows the file and error
