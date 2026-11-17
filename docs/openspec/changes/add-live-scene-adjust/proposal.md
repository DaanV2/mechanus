# Proposal

## Why

Milestone M4 (The room comes alive). The GM needs to tweak lights and volume mid-scene without editing the scene.

## What Changes

- `SceneService.Adjust` to override a target's instruction live
- GM view controls for light and volume
- Adjustments stored as scene state overrides

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scene-control`: adds live adjustments

## Impact

- `scenes/v1`, scene state overrides, GM view
