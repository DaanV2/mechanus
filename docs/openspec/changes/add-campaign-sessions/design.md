# Design

## Context

Scene state is currently saved per scene only.

## Decisions

- A session has no start / end lifecycle
- Only the latest scene state per campaign + scene is kept, overwritten on exit; no history

## Risks / Trade-offs

- Existing scene state has no campaign → migrate it to a default campaign or drop it; decide in the migration task
