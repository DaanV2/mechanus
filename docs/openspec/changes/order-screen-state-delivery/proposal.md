# Proposal

## Why

Parking lot (no milestone). `ScreenHandler.Broadcast` writes each message in its own goroutine, so two quick changes (a GM switching variant twice) can arrive out of order and leave the TV on the older state. Initial setup replies can overtake a newer state the same way. Seen in the code, not yet in practice.

## What Changes

- Deliver screen messages to each connection in the order they were produced

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `screen-connection`: guarantees per-connection message order

## Impact

- `ScreenHandler` broadcast and initial setup paths
