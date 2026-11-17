# Spec Delta

## ADDED Requirements

### Requirement: Spotify music
The system SHALL play the playlist or track named by a scene's music instruction on Spotify, and apply its volume and stop instructions.

#### Scenario: Scene starts
- **WHEN** a scene names a Spotify playlist
- **THEN** the playlist starts on the configured player

#### Scenario: Not linked
- **WHEN** no Spotify account is linked
- **THEN** the device status shows it as not configured and the scene still activates
