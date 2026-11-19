# Scene and room files

Scenes and the room are authored configuration: files you write (or, later, the in-app editor writes) that the server only reads. Everything that changes during play is scene state and lives in the database. See [scenes.md](../architecture/scenes.md) for the model behind these files.

## Where they live

```text
<app config dir>/
├── rooms/room.yaml           The room: devices and the targets they serve
├── scenes/
│   ├── tavern.yaml           A scene config, its ID is the file name
│   └── tavern/               The scene's assets, see battle-maps.md
└── shared/                   Assets shared between scenes
```

`mechanus config paths` prints the app config dir on your machine, see [config-file.md](./config-file.md).

## Formats

A file is YAML (`.yaml` or `.yml`) or JSON (`.json`), picked by its extension. Both use the same field names, so any example below can be written either way. Files are read from the top of the `scenes` and `rooms` folders only; sub folders hold assets and files starting with `.` are ignored.

- The scene ID is the file name without extension (`tavern.yaml` is scene `tavern`). IDs may contain letters, digits, `-` and `_`. An `id` field may be left out; when present it must match the file name.
- The room is always `rooms/room.yaml` (or `.yml` / `.json`).
- Only one file per ID: `tavern.yaml` next to `tavern.json` is reported as an error.
- `.dat` files from older versions are no longer read and are reported as an error. Rename them to `.json` and add `"version": 1`.

## Versions

Every file starts with the version of its format:

```yaml
version: 1
```

The current version of both scene and room files is **1**. When the format changes, the version goes up and the server migrates older files in memory when it loads them; it never rewrites your files. A file with a version newer than the server knows, or without a version, is reported as invalid.

## Scene file

```yaml
version: 1
name: Tavern
description: The Prancing Goat, ground floor
campaign_id: ""                     # Optional campaign the scene belongs to
default_variant: day                # Optional, variant applied the first time the scene loads

entities:                           # Who / what is present when the scene starts
  - { id: barkeep, kind: npc, name: Barkeep }        # kind: player, npc, enemy or object
  - { id: spy, kind: enemy, name: Spy, hidden: true } # hidden from screens until revealed

instructions:                       # Base instructions per target
  screens:
    table: { view: battlemap, asset: maps/tavern.map.json, overlays: [{ kind: entities }] }
  lights:
    ambient: { on: true, color: "#ffaa55", brightness: 0.6 }
  music:
    music: { action: play, source: "spotify:playlist:...", volume: 0.4 }
  custom:
    fog-machine: { payload: { on: true } }    # Any JSON / YAML, passed through untouched

variants:                           # Override the base instructions per target
  - id: day
    name: Day
    tags: { time_of_day: day }
    instructions: {}
  - id: night
    name: Night
    tags: { time_of_day: night }
    instructions:
      screens:
        table: { view: battlemap, asset: maps/tavern-night.map.json, overlays: [{ kind: entities }] }
      lights:
        ambient: { on: true, color: "#3344aa", brightness: 0.2 }
```

Durations (`fade` on lights, `cross_fade` on music) are in nanoseconds: `fade: 2000000000` is two seconds.

## Room file

```yaml
version: 1
name: Living room
devices:                            # Device ID -> device
  table-tv: { id: table-tv, name: Table TV, kind: screen, integration: screen }
  ceiling: { id: ceiling, name: Ceiling light, kind: light, integration: mqtt }
bindings:                           # Target -> device IDs
  table: [table-tv]
  ambient: [ceiling]
```

Paired screens are registered in the database and can be bound to targets from the app; those bindings are added to the ones in the file. See [integrations.md](./integrations.md) for each integration's `config`.

## Editing files

The server reads the `scenes` and `rooms` folders when it starts; restart it after editing a file. A file that is invalid (syntax error, unknown version, wrong ID, ...) is skipped and the GM view lists the file with the error.

Battle map files (`*.map.json`) are read every time a screen is instructed, so they need no restart.

## Example

[`docs/examples/battlemap`](../examples/battlemap) holds a working room and scene, mounted by `docker compose watch`.
