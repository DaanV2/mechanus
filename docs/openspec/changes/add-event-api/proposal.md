# Proposal

## Why

Milestone M4 (The room comes alive). External software (stream overlays, home automation, custom props) should be able to follow scene events and send events back without MQTT.

## What Changes

- JSON WebSocket event API using the same event schema as MQTT
- Admin-created, revocable, scoped API tokens (JWT + JTI)
- Clients subscribe to scene events and can send events back

## Capabilities

### New Capabilities

- `event-api`: the third-party event WebSocket and its API tokens

### Modified Capabilities

None.

## Impact

- new WebSocket route, token model and admin UI, docs/api/
