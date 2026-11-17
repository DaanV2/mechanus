# Design

## Context

Security checks live in the message handlers (see `move-security-checks-into-handlers`), so the GM checks go there too. Today campaigns have an owner (`Campaign.OwnerID`) that stands in for a GM.

## Decisions

- `gm` is a role next to `admin` and `user`; admins keep every GM permission
- Invites given out by a GM still create `user` accounts; a GM cannot hand out the `gm` or `admin` role

## Risks / Trade-offs

- Existing users lose scene control on upgrade → grant `gm` to campaign owners in a migration, or document the manual step

## Open Questions

- Is GM per campaign (the user who runs that campaign) or global (a role on the account)? Decide before writing tasks; it changes the campaign and scene checks.
- How screens are owned for the per-screen access check: by the room (one instance, one room) with the GM allowed everything, or per campaign session.
