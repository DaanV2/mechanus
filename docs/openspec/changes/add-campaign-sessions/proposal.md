# Proposal

## Why

Milestone M5 (Sessions & campaigns) — also covers the M2 item "scene state keyed by campaign + scene". A campaign's game should continue exactly where it left off, across server restarts and campaign switches. Today the active scene is forgotten on restart and scene state is not tied to a campaign.

## What Changes

- One never-ending session per campaign holding the active scene and play state
- Scene state keyed by campaign + scene (scenes stay campaign-independent)
- Remember the active scene across server restarts
- Switching campaign switches session; returning resumes it

## Capabilities

### New Capabilities

- `campaign-sessions`: the continuous play session of a campaign

### Modified Capabilities

- `scene-control`: scene state is stored per campaign + scene

## Impact

- scene state storage (migration of existing state), session model, GM view campaign switcher
