# Spec Delta

## ADDED Requirements

### Requirement: Initiative overlay
The system SHALL support an `initiative` overlay showing the initiative order with the current turn highlighted.

#### Scenario: Turn advances
- **WHEN** the GM advances the turn
- **THEN** the overlay highlights the new current entity

### Requirement: Entity overlay on all views
The system SHALL support the `entities` overlay, with portraits, on every view kind, not only battle maps.

#### Scenario: Scenery with entities
- **WHEN** a scenery instruction enables the `entities` overlay
- **THEN** portraits and names of visible entities are shown over the scenery
