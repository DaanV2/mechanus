# Spec Delta

## ADDED Requirements

### Requirement: Matter lights
The system SHALL apply a scene's light instructions (power, brightness, colour, fade) to lights reached through Matter.

#### Scenario: Scene sets a colour
- **WHEN** a scene instructs a Matter light to red at 40%
- **THEN** the physical light turns red at 40% brightness

#### Scenario: Integration unreachable
- **WHEN** Matter cannot be reached
- **THEN** the device status shows it as unreachable and other devices are unaffected
