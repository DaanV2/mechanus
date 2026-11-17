# Design

## Context

MQTT already publishes scene events (see `device-drivers`). The event API mirrors that schema.

## Decisions

- JSON WebSocket only
- Same event schema as MQTT
- API tokens are JWT + JTI, created and revoked by admins

## Open Questions

- Exact scopes. Candidates: `events:read` (subscribe), `events:write` (custom events in the GM view), `scenes:control` (activate / switch variant / adjust) — roadmap TBD
