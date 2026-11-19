# Spec Delta

## REMOVED Requirements

### Requirement: Live reload
**Reason**: Watching is unreliable on Docker Desktop bind mounts, and re-applying an edited active scene mid-session disrupts play. The scene editor will save through the server instead.
**Migration**: Restart the server after editing scene or room files.

### Requirement: Invalid files keep the last good version
**Reason**: Without reloading, a file has no earlier version to fall back to.
**Migration**: See "Invalid files are skipped and reported".

## ADDED Requirements

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
