# Spec Delta

## Purpose

Lets third-party software subscribe to scene events and send events back over a WebSocket, authenticated with scoped API tokens.

## ADDED Requirements

### Requirement: API tokens
The system SHALL let admins create and revoke scoped API tokens for the event API.

#### Scenario: Revoked token
- **WHEN** an admin revokes an API token
- **THEN** connections using it are closed and new ones refused

### Requirement: Subscribe to scene events
The system SHALL stream scene events as JSON, in the same schema as the MQTT scene events, to clients holding a token with the read scope.

#### Scenario: Variant switch
- **WHEN** the GM switches variant
- **THEN** subscribed clients receive the scene event

### Requirement: Send events
The system SHALL accept events from clients holding a token with a write scope and reject them otherwise.

#### Scenario: Missing scope
- **WHEN** a read-only token sends an event
- **THEN** the event is rejected
