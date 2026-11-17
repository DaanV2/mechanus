# Design

## Context

See docs/architecture/authentication.md ("Where security checks go"). The interceptor checks a route policy before the handler runs, so the handler cannot apply per-resource rules.

## Decisions

- Entry function validates (auth, role, ownership, input) then calls the private variant; internal callers use the private variant directly

## Open Questions

- Whether the screen WebSocket's `http.RequirePolicy` moves into the handler as well (roadmap TBD)
