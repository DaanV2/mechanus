# Spec Delta

## ADDED Requirements

### Requirement: Device view keeps the screen awake
The client SHALL keep the screen awake while the device view is open, using the Screen Wake Lock API in a secure context and a fallback over plain HTTP, and SHALL re-acquire it when the page becomes visible again.

#### Scenario: Plain HTTP TV
- **WHEN** a TV shows the device view over plain HTTP for an hour without input
- **THEN** the screen does not dim or sleep

#### Scenario: Tab hidden and shown
- **WHEN** the device view is hidden and then shown again
- **THEN** the screen is kept awake again
