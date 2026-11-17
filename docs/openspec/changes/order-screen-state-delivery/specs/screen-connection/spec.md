# Spec Delta

## ADDED Requirements

### Requirement: Ordered delivery per connection
The system SHALL deliver screen messages to each connection in the order they were produced, including initial setup replies.

#### Scenario: Two quick variant switches
- **WHEN** the GM switches variant twice in quick succession
- **THEN** every screen ends on the state of the second switch
