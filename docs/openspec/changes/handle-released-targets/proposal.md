# Proposal

## Why

Milestone M4 (The room comes alive). When a scene change leaves a target without an instruction (`Transition.Released`), its device keeps whatever it was doing; there is no defined behaviour.

## What Changes

- Defined behaviour for released targets: reset, stop, or keep

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scene-control`: defines what happens to targets a new scene or variant no longer instructs

## Impact

- scene transitions, output drivers
