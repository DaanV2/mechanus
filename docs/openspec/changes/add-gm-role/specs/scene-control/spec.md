# Spec Delta

## ADDED Requirements

### Requirement: Scene control requires the GM role
The system SHALL only let users with the `gm` or `admin` role activate scenes, switch variants, reset scenes and use the GM entity controls.

#### Scenario: Player tries to switch variant
- **WHEN** a user without the `gm` or `admin` role switches the active scene's variant
- **THEN** the system responds `PermissionDenied`
