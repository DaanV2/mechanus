# Proposal

## Why

Milestone M2 (Scene authoring). GMs already have maps made in other tools. `mechanus maps import` is a stub today.

## What Changes

- Import Universal VTT (`.dd2vtt`) files into a battle map
- Import Dungeondraft maps
- Implement the `mechanus maps import` command

## Capabilities

### New Capabilities

- `map-import`: converting maps from other tools into Mechanus battle maps

### Modified Capabilities

None.

## Impact

- `server/cmd` maps command
- battle map writer
