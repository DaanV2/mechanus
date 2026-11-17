# Spec Delta

## ADDED Requirements

### Requirement: Released targets
When a scene activation or variant switch leaves a previously instructed target without an instruction, the system SHALL apply the released-target behaviour configured for that device.

#### Scenario: Music not in the new scene
- **WHEN** the new scene has no music instruction
- **THEN** the music device applies its released-target behaviour instead of silently keeping the old track
