# Spec Delta

## ADDED Requirements

### Requirement: Philips Hue lights
The system SHALL apply a scene's light instructions (power, brightness, colour, fade) to lights reached through Philips Hue.

#### Scenario: Scene sets a colour
- **WHEN** a scene instructs a Philips Hue light to red at 40%
- **THEN** the physical light turns red at 40% brightness

#### Scenario: Integration unreachable
- **WHEN** Philips Hue cannot be reached
- **THEN** the device status shows it as unreachable and other devices are unaffected
