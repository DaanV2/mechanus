# Spec Delta

## ADDED Requirements

### Requirement: iTunes / Apple Music music
The system SHALL play the playlist or track named by a scene's music instruction on iTunes / Apple Music, and apply its volume and stop instructions.

#### Scenario: Scene starts
- **WHEN** a scene names a iTunes / Apple Music playlist
- **THEN** the playlist starts on the configured player

#### Scenario: Not linked
- **WHEN** no iTunes / Apple Music account is linked
- **THEN** the device status shows it as not configured and the scene still activates
