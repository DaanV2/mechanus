# Spec Delta

## ADDED Requirements

### Requirement: Per-screen access for users
The system SHALL let GMs and admins open every screen and let other users open only the screens they are allowed to see, answering `403` otherwise.

#### Scenario: Player opens the GM's screen
- **WHEN** a player requests a screen they are not allowed to see
- **THEN** the system responds `403`
