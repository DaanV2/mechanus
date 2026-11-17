# Spec Delta

## ADDED Requirements

### Requirement: Touch modes
The system SHALL apply a per-device `touch` setting: `none` ignores touch, `view` (the default) allows pan, zoom, ping and measure, and `control` also moves tokens as the configured role or user.

#### Scenario: View mode
- **WHEN** a player drags a token on a `view` mode table screen
- **THEN** the token does not move

#### Scenario: Control mode
- **WHEN** a table screen is in `control` mode as a player
- **THEN** tokens that player controls can be moved by touch

### Requirement: Touch lock
The system SHALL let the GM disable touch input on screens live.

#### Scenario: Lock touch
- **WHEN** the GM locks touch
- **THEN** touch input on screens is ignored until unlocked
