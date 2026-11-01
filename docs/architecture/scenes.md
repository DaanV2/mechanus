# Scenes & Outputs

Scenes are the core of Mechanus. The GM prepares scenes, and activating one makes the whole room follow: the table TV shows the battle map, the wall screen shows the scenery, the lights dim to candlelight and the tavern playlist starts. This document describes the target model and which parts of it exist today.

## Goals

- A **scene** records which entities are present (players, NPCs, enemies, objects).
- A scene has **variants** the GM built (time of day, weather, combat, ...) and can transition between them.
- A scene **instructs devices** what to do:
  - **Screens** show a _view_: battle map, scenery, information board, or a custom view, optionally with overlays such as an initiative tracker or the entities on screen.
  - **Lights** (Hue, Matter, Home Assistant, custom) get colour / brightness / effect instructions.
  - **Music** (Spotify, iTunes, custom) gets play / pause / volume instructions.
  - **Custom devices** get raw payloads over an event bus (MQTT or similar).

## Configuration vs. state

Everything in this model is either **configuration** or **state**, and the two are kept in separate types and separate storage.

|            | Configuration                                                                        | State                                                                                         |
| ---------- | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| What       | What the GM prepared: scenes, variants, starting entities, rooms, devices            | What happened during play: active variant, revealed / removed / spawned entities, live tweaks |
| Types      | `scenes.Config`, `scenes.Variant`, `scenes.Entity`, `devices.Room`, `devices.Device` | `scenes.State`, `scenes.EntityState`                                                          |
| Stored in  | **File storage** (app config folder)                                                 | **Database** (key-value table)                                                                |
| Written by | The GM / tooling, outside of play. The server only reads it.                         | The server, when a scene is loaded and when it is exited                                      |
| Adapter    | `infrastructure/scenestore.ConfigStore`                                              | `infrastructure/scenestore.StateStore`                                                        |

A **loaded scene** (`scenes.Scene`) is a `Config` plus its `State`. It computes the effective instructions and entities from both and never modifies the `Config`.

State only records what differs from or adds to the configuration. That keeps the configuration the source of truth: if the GM renames an entity or tweaks a variant in the config, a saved state picks it up on the next load. Entities added to the config after a state was saved are added on load; a saved variant that no longer exists falls back to the default.

### Lifecycle

```ascii
                    file storage                                      database
              ┌──────────────────────┐                        ┌──────────────────────┐
              │ scenes/<id>.dat      │                        │ state/scene/<id>     │
              │ rooms/<id>.dat       │                        │                      │
              └──────────┬───────────┘                        └───▲──────────┬───────┘
                         │ read only                       save   │          │ restore (or none
                         ▼                                        │          ▼  → NewState(config))
  Activate(scene) ──► Config ──────────────► scenes.Load(Config, State) ──► active Scene
                                                                          │   │
                                         SetVariant / Adjust / entity ────┘   │ Activate(other) or Exit()
                                         changes mutate State only            ▼
                                                                    save State, deactivate
```

1. **Load** (`SceneService.Activate`): read the `Config` from file storage, restore its `State` from the database (or create a fresh one from the config), save it, instruct the devices.
2. **Play**: switching variants, live adjustments and entity changes only mutate the `State`.
3. **Exit** (activating another scene, or `SceneService.Exit` on shutdown): save the `State` to the database. If saving fails, the switch is aborted so no state is lost.

## Concepts

```ascii
                     configuration, shared by campaigns               configuration, per room
 ┌──────────────────────────────────────────────────────┐      ┌──────────────────────────────┐
 │ Scene config "tavern"                                │      │ Room "living room"           │
 │   entities: barkeep (npc), spy (enemy, hidden), ...  │      │                              │
 │   instructions (base):                               │      │  target   →  devices         │
 │     screens.table   → battlemap tavern (custom)      │      │  table    →  TV (screen)     │
 │     screens.wall    → scenery  tavern-day.png        │      │  wall     →  monitor (screen)│
 │                       + initiative tracker, entities │      │                              │
 │     lights.ambient  → #ffcc88 80%                    │      │  ambient  →  hue-1, hue-2    │
 │     music.music     → play spotify:playlist:tavern   │      │  music    →  spotify-speaker │
 │   variants:                                          │      │  fog      →  mqtt fog-machine│
 │     night → screens.wall, lights.ambient overridden  │      │                              │
 └──────────────────────────┬───────────────────────────┘      └───────────────┬──────────────┘
                            │ + State (variant, overrides, entities)           │
                            ▼                                                  │
                     Scene.Instructions()                                      │
                            │ NewTransition(prev, next)                        │ DevicesFor(target, kind)
                            ▼                                                  ▼
                     Transition.Changes ─────────────────────────────► SceneService.dispatch
                                                                               │ by device.Integration
                                         ┌──────────────┬──────────────┬───────┴──────┬──────────────┐
                                         ▼              ▼              ▼              ▼              ▼
                                   screen driver    hue driver    spotify driver  mqtt driver   ... driver
                                   (WebSocket)
```

| Concept          | Kind   | What it is                                                                                                                                        | Code                              |
| ---------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| **Scene config** | config | A place or moment in the game. Starting entities, base instructions, variants, default variant. Not tied to a campaign: any campaign can use it, state is kept per campaign. | `engine/scenes.Config`            |
| **Entity**       | config | Something present in the scene: `player`, `npc`, `enemy`, `object`. Can start hidden. Optionally links to a persisted character.                  | `engine/scenes.Entity`            |
| **Variant**      | config | An alternative state of a scene (night, rain, combat). Overrides base instructions per target. Has free-form tags (`time_of_day: night`).         | `engine/scenes.Variant`           |
| **Instructions** | config | What every output should do, keyed by **target** per device kind (`Screens`, `Lights`, `Music`, `Custom`).                                        | `engine/scenes.Instructions`      |
| **Target**       | config | A logical output name chosen by the GM (`table`, `wall`, `ambient`, `music`). Scenes only ever talk to targets, never to physical devices.        | map keys in `Instructions`        |
| **Room**         | config | The physical setup. Lists devices and binds targets to them. The same scene can be played in a different room by changing only the bindings.      | `engine/devices.Room`             |
| **Device**       | config | A physical or virtual output with a `Kind` (`screen`, `light`, `music`, `custom`) and an `Integration` (`screen`, `hue`, `spotify`, `mqtt`, ...). | `engine/devices.Device`           |
| **Scene state**  | state  | Active variant, per-entity status (hidden / removed), spawned entities, live instruction overrides.                                               | `engine/scenes.State`             |
| **Scene**        | both   | A loaded scene: config + state. Computes effective instructions and present entities.                                                             | `engine/scenes.Scene`             |
| **Transition**   | —      | What changes for the devices between two sets of instructions: full instructions, changes to send, released targets.                              | `engine/scenes.Transition`        |
| **Driver**       | —      | Infrastructure adapter that delivers an instruction to one device of its integration.                                                             | `application.OutputDriver` (port) |

### Why targets instead of devices

Scenes are prepared ahead of time and shared between sessions; hardware changes (a TV is swapped, a light is added, the game moves to another house). Keeping scenes device-agnostic means the only thing that changes is the room's bindings. A target can be bound to several devices (all lights in `ambient`), and a device can serve several targets.

A target name is shared across kinds: `table` can have a screen instruction _and_ a light instruction, and each is only sent to devices of the matching kind bound to `table`.

## Rules

### Effective instructions

`Scene.Instructions()` layers three sources, later ones replacing earlier ones per target:

1. The config's base instructions.
2. The active variant's instructions (`State.VariantID`; empty means base only).
3. The state's live overrides (`State.Overrides`), set by `Scene.Adjust` — e.g. the GM turning the music down mid-scene.

Replacement is per target and kind, not per field: an override for `lights.ambient` replaces the whole light instruction for that target. Targets not mentioned keep the instruction from the layer below.

`Config.Resolve(variantID)` gives layers 1–2 only. Empty `variantID` → `Config.DefaultVariant`; unknown variant → `ErrVariantNotFound`.

Switching variant (`Scene.SetVariant`) **clears the live overrides**: a variant is a deliberate change of the whole scene's look.

### Entities

The config lists the entities present when the scene starts (`Entity.Hidden` = starts hidden). During play the state tracks, per entity ID, whether it is hidden or removed (`State.SetHidden`, `State.Remove`), and holds entities added mid-scene (`State.Spawn`). `Scene.Entities()` returns configured + spawned entities that are not removed, with their current hidden status; `Scene.VisibleEntities()` filters out hidden ones.

### Transitions

`scenes.NewTransition(from, prev, to, next)` compares two sets of effective instructions. Scene switches, variant switches, and live adjustments are all transitions.

- `Changes` contains only instructions that are new or differ, so a variant that only changes the lights does not reload the battle map.
- `Released` lists targets that were instructed before but no longer are. What to do with them (blank the screen, stop the music, leave as is) is **not decided yet** — `SceneService` currently ignores them.
- Fading / cross-fading is part of the instruction (`LightInstruction.Fade`, `MusicInstruction.CrossFade`), so each integration handles it natively.

### SceneService

`application.SceneService` owns the active scene and talks to the stores through ports:

| Port               | Backed by    | Used for                                          |
| ------------------ | ------------ | ------------------------------------------------- |
| `SceneConfigStore` | file storage | `GetSceneConfig(sceneID)`                         |
| `RoomConfigStore`  | file storage | `CurrentRoom()` — devices and target bindings     |
| `SceneStateStore`  | database     | `LoadSceneState` / `SaveSceneState`               |
| `OutputDriver`     | integrations | `Apply(scene, device, instruction)`, one per integration |

| Method                         | Effect                                                                  |
| ------------------------------ | ----------------------------------------------------------------------- |
| `Activate(sceneID, variantID)` | Load (see [Lifecycle](#lifecycle)). Same scene → only switches variant. |
| `SetVariant(variantID)`        | Switch variant of the active scene, clears live overrides.              |
| `Adjust(overrides)`            | Apply live overrides to the active scene.                               |
| `Exit()`                       | Save state, deactivate. Intended for `BeforeShutdown`.                  |
| `Active()`                     | Active scene + variant.                                                 |

Each change computes a transition and dispatches `Changes`: for every changed instruction, find the devices bound to its target with the matching kind and hand each to the driver registered for `device.Integration`. Device errors (including `ErrNoDriver`) are joined and returned; other devices are still instructed.

## Storage layout

`infrastructure/scenestore` builds on the generic `infrastructure/storage` providers:

| Data          | Provider                                     | Location                                  |
| ------------- | -------------------------------------------- | ----------------------------------------- |
| Scene configs | `NamedFileStorage[*scenes.Config]("scenes")` | `<app config dir>/scenes/<scene id>.dat`  |
| Rooms         | `NamedFileStorage[*devices.Room]("rooms")`   | `<app config dir>/rooms/<room id>.dat`    |
| Scene states  | `NamedDBStorage[*scenes.State](db, "scene")` | `key_values` row `state/scene/<scene id>` |

Files and rows contain JSON. Scene and room IDs are used as file names, so `ConfigStore` only accepts `[A-Za-z0-9_-]` IDs and checks that a file's `id` matches its name.

## Outputs

### Screens — views

A `ScreenInstruction` selects a view; the client renders it. Screens are browser clients connected over the screen WebSocket (see [communication.md](./communication.md)).

| View        | Purpose                                                        | Typical `Asset`                 |
| ----------- | -------------------------------------------------------------- | ------------------------------- |
| `battlemap` | Tactical grid with entities, fog of war, tokens                | Mechanus' own battle map format |
| `scenery`   | Atmospheric image or video backdrop                            | image / video                   |
| `infoboard` | Handouts, initiative order, quest log, text                    | document / none                 |
| `custom`    | Client component chosen by `Component`, configured by `Params` | anything                        |
| `splash`    | Title card / loading screen                                    | image                           |
| `blank`     | Black screen                                                   | —                               |

**Battle maps** use a custom, Mechanus-specific format rather than an existing VTT format, so the map can carry everything the battle map view needs. A battle map is grid settings plus an ordered list of typed layers: v0 has image layers and entity layers (including instance groups — one asset drawn at many positions); walls, lights, sound and other layer kinds are added when the feature that uses them lands. Importers for other formats (e.g. the Universal VTT parser in `infrastructure/vttrpg`) would convert into it rather than being rendered directly. The map is configuration; anything that moves during play (token positions, revealed fog) belongs in state.

#### Overlays

Any view can carry `Overlays`: widgets drawn on top of it. This lets a scenery screen also show the initiative tracker, or portraits of the entities in the scene, without needing a dedicated view per combination.

| Overlay      | Purpose                                                        |
| ------------ | -------------------------------------------------------------- |
| `initiative` | Initiative / turn order tracker                                |
| `entities`   | Names / portraits of the scene's visible (non-hidden) entities |
| `custom`     | Client component chosen by `Component`, configured by `Params` |

Overlays are part of the screen instruction, so a variant (e.g. `combat`) can add the initiative tracker to the wall screen by overriding that target's instruction, and only that screen is re-instructed.

The `engine/screens` package (`ScreenHandler`, `ScreenManager`, `state.ScreenState`) is the transport for the `screen` integration: the screen driver (`infrastructure/outputs/screen`) translates a `ScreenInstruction` into `ScreenState` and broadcasts it to the screen's listeners. There is one `ScreenHandler` per screen device, so the WebSocket `screenId` is the device ID and routing by role or device is answered by the room's bindings. Entities are not sent yet; they need the battle map format for positions.

### Lights

`LightInstruction`: `On`, `Color` (hex), `Brightness` (0–1), `Effect` (integration specific, e.g. `candle`), `Fade`. Planned integrations: Philips Hue, Matter, Home Assistant, custom.

### Music

`MusicInstruction`: `Action` (`play`/`pause`/`stop`), `Source` (integration specific URI, e.g. `spotify:playlist:...`), `Volume` (0–1), `Shuffle`, `Loop`, `CrossFade`. Planned integrations: Spotify, iTunes / Apple Music, custom.

### Custom devices

`CustomInstruction`: `Topic` + opaque JSON `Payload`, passed through untouched. Planned: scenes only carry the payload per target; the topic comes from the device config (default `mechanus/devices/<device-id>/set`) and the payload is wrapped in an envelope `{scene, variant, target, payload}` (see [roadmap M4](../roadmap.md#m4--the-room-comes-alive)). Intended for MQTT first (fog machines, rumble packs, Arduino props), but any event bus can implement a driver.

## Code map

```ascii
server/
├── engine/
│   ├── devices/            Device, Kind, Room (target → device bindings), integration identifiers   [config]
│   ├── scenes/
│   │   ├── config.go       Config, Entity, Variant                                                    [config]
│   │   ├── state.go        State, EntityState                                                         [state]
│   │   ├── scene.go        Scene = Config + State, Load                                               [both]
│   │   ├── instructions.go Instructions and the per-kind instruction types
│   │   └── transition.go   Ref, Transition, NewTransition
│   └── screens/            Screen WebSocket fan-out — the transport for the "screen" integration
├── application/
│   └── scene_service.go    SceneService + ports: SceneConfigStore, RoomConfigStore, SceneStateStore, OutputDriver
└── infrastructure/
    ├── assets/             Resolves scene / shared asset references to hashed URLs, serves /assets/
    ├── outputs/screen/     Screen driver: ScreenInstruction → ScreenState on the device's screen
    ├── scenestore/         ConfigStore (file storage), StateStore (database)
    └── storage/            Generic file / database storage providers
```

## Status

**Implemented**: the domain model above with the config / state split, variant resolution, live overrides, entity status tracking, transition change detection, `SceneService` with load / exit persistence, `scenestore` adapters for file storage and the database, and tests for each. `SceneService` is wired in `components` with the instance's single room (`rooms/room.dat`, `scenestore.RoomID`; a missing room file is an empty room) and saves the active scene on `BeforeShutdown`. The screen driver (`infrastructure/outputs/screen`) turns a `ScreenInstruction` into screen state (view, asset URL, grid shown for battle maps) and pushes it to the device's screen; screens are keyed by device ID. `scenes/v1.SceneService` (ConnectRPC) lists scenes, activates them, switches variant and returns the active scene. Scene and shared assets are served under `/assets/` with a content hash (`infrastructure/assets`).

**Not implemented yet**:

- The battle map format itself (the screen driver sends no layers, grid settings or entities yet).
- Authoring: no tooling or API to create / edit scene and room configs (files have to be written by hand).
- Remembering which scene was active across a server restart.
- No drivers other than the screen driver: Hue, Matter, Home Assistant, Spotify, iTunes, MQTT.
- The scene control API has no `Adjust` or entity (reveal / spawn) calls yet.
- Handling of `Transition.Released`.
- Device registration, per-route auth policies and pairing (designed in [authentication.md](./authentication.md#device-authentication)).

## Extending

**Adding something to a scene**: first decide whether it is configuration (prepared before play, read from files) or state (changes during play, saved to the database). Configuration goes on `Config` / `Variant`; state goes on `State`, storing only the difference from the config. Never mutate a `Config` at runtime.

**Adding an integration driver**

1. Implement `application.OutputDriver` in `server/infrastructure/outputs/<integration>/` (type-switch on the `scenes.Instruction`; ignore kinds you do not handle).
2. Read integration specific settings from `Device.Config` (bridge address, entity ID, MQTT topic, ...). Global settings (credentials, broker URL) go in a Viper config set.
3. Add an identifier constant in `engine/devices/kind.go` if it is new.
4. Register the driver when constructing `SceneService` in `components`.

**Adding a view kind**: add a `ViewKind` constant in `engine/scenes/instructions.go`, then teach the screen driver and the client renderer about it. Prefer `ViewCustom` + `Component` for one-off views.

**Adding a device kind**: add a `devices.Kind`, an instruction type implementing `scenes.Instruction`, a map on `Instructions`, and extend `Merge`, `Changes`, `Released`, `IsEmpty`, and `All`.

## Open questions

- Should variants be able to add / remove / reveal entities (an ambush at night), or only change outputs?
- Released targets: reset to an idle state, or keep the last instruction?
- Should transitions between specific variants be restricted or have their own timing, or is per-instruction fade enough?
- Player views: does a player's personal screen get a target per player, or a role-based view derived from the battle map?
- Should a scene's state be resettable to its config (start the scene over), and should old states be kept per session?
