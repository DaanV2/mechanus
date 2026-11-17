# Proposal

## Why

Milestone M2 (Scene authoring). A GM should be able to create and edit scenes, variants and the room setup in the app instead of writing files by hand. The in-app editor is the primary authoring path; hand-edited files remain supported.

## What Changes

- Scene editor: entities, base instructions per target, variants
- Room editor: devices and target bindings
- Preview a scene or variant without activating it
- Editor saves to the same config files that can be hand-edited

## Capabilities

### New Capabilities

- `scene-authoring`: creating and editing scenes, variants and the room from the app, and previewing them

### Modified Capabilities

None.

## Impact

- New authoring RPCs in `scenes/v1`
- GM client editor pages
- Depends on `add-versioned-config-files` for the file format
