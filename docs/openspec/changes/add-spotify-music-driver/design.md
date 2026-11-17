# Design

## Context

Mechanus is LAN-only and usually has no public HTTPS URL.

## Decisions

None yet; settle them before writing tasks.

## Open Questions

- How to complete Spotify OAuth without a public callback URL: project-hosted static HTTPS relay page (PKCE, user's own client ID), loopback redirect and paste the URL back, delegate to Home Assistant, or require TLS (roadmap TBD). Resolve before writing tasks for linking.
