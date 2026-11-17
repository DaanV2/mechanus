# Spec Delta

## Purpose

Converts maps made in other tools into Mechanus battle map files so GMs can reuse existing maps.

## ADDED Requirements

### Requirement: Import Universal VTT
The system SHALL convert a Universal VTT (`.dd2vtt`) file into a battle map file with its image and grid settings.

#### Scenario: Import a dd2vtt
- **WHEN** the GM runs `mechanus maps import cave.dd2vtt`
- **THEN** a `*.map.json` and its image are written to the target scene folder

### Requirement: Import Dungeondraft
The system SHALL convert a Dungeondraft export into a battle map file.

#### Scenario: Import Dungeondraft
- **WHEN** the GM imports a Dungeondraft map
- **THEN** a battle map file with the map image and grid is created
