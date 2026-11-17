# Spec Delta

## ADDED Requirements

### Requirement: Live adjustments
The system SHALL let the GM adjust a target's instruction (for example brightness or volume) during play, storing it as a scene state override until the scene is reset.

#### Scenario: Dim the lights
- **WHEN** the GM lowers the brightness of the `ceiling` target
- **THEN** the light dims and stays dimmed when the scene is re-activated
