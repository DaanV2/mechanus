# Scenes & Outputs

Scenes are the core of Mechanus. The GM prepares scenes, and activating one makes the whole room follow: the table TV shows the battle map, the wall screen shows the scenery, the lights dim to candlelight and the tavern playlist starts. This document describes the target model and which parts of it exist today.

## Goals

- A **scene** records which entities are present (players, NPCs, enemies, objects).
- A scene has **variants** the GM built (time of day, weather, combat, ...) and can transition between them.
- A scene **instructs devices** what to do:
  - **Screens** show a _view_: battle map, scenery, information board, or a custom view.
  - **Lights** (Hue, Matter, Home Assistant, custom) get colour / brightness / effect instructions.
  - **Music** (Spotify, iTunes, custom) gets play / pause / volume instructions.
  - **Custom devices** get raw payloads over an event bus (MQTT or similar).

## Concepts

```ascii
                        authored per campaign                         configured per room
 ┌──────────────────────────────────────────────────────┐      ┌──────────────────────────────┐
 │ Scene "tavern"                                       │      │ Room "living room"           │
 │   entities: barkeep (npc), spy (enemy, hidden), ...  │      │                              │
 │   instructions (base):                               │      │  target   →  devices         │
 │     screens.table   → battlemap tavern.dd2vtt        │      │  table    →  TV (screen)     │
 │     screens.wall    → scenery  tavern-day.png        │      │  wall     →  monitor (screen)│
 │     lights.ambient  → #ffcc88 80%                    │      │  ambient  →  hue-1, hue-2    │
 │     music.music     → play spotify:playlist:tavern   │      │  music    →  spotify-speaker │
 │   variants:                                          │      │  fog      →  mqtt fog-machine│
 │     night → screens.wall, lights.ambient overridden  │      │                              │
 └──────────────────────────┬───────────────────────────┘      └───────────────┬──────────────┘
                            │ Resolve(variant) + Plan(from, to)                │ DevicesFor(target, kind)
                            ▼                                                  ▼
                     Transition.Changes ─────────────────────────────► SceneService.dispatch
                                                                               │ by device.Integration
                                         ┌──────────────┬──────────────┬───────┴──────┬──────────────┐
                                         ▼              ▼              ▼              ▼              ▼
                                   screen driver    hue driver    spotify driver  mqtt driver   ... driver
                                   (WebSocket)
```

| Concept         | What it is                                                                                                                                                         | Code                                   |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------- |
| **Scene**       | A place or moment in the game. Owns entities, base instructions, and variants. Belongs to a campaign.                                                              | `engine/scenes.Scene`                  |
| **Entity**      | Something present in the scene: `player`, `npc`, `enemy`, `object`. Can be hidden from players. Optionally links to a persisted character.                          | `engine/scenes.Entity`                 |
| **Variant**     | An alternative state of a scene (night, rain, combat). Overrides base instructions per target. Has free-form tags (`time_of_day: night`).                          | `engine/scenes.Variant`                |
| **Instructions**| What every output should do, keyed by **target** per device kind (`Screens`, `Lights`, `Music`, `Custom`).                                                          | `engine/scenes.Instructions`           |
| **Target**      | A logical output name chosen by the GM (`table`, `wall`, `ambient`, `music`). Scenes only ever talk to targets, never to physical devices.                           | map keys in `Instructions`             |
| **Room**        | The physical setup. Lists devices and binds targets to them. The same scene can be played in a different room by changing only the bindings.                        | `engine/devices.Room`                  |
| **Device**      | A physical or virtual output with a `Kind` (`screen`, `light`, `music`, `custom`) and an `Integration` (`screen`, `hue`, `spotify`, `mqtt`, ...).                   | `engine/devices.Device`                |
| **Driver**      | Infrastructure adapter that delivers an instruction to one device of its integration.                                                                              | `application.OutputDriver` (port)      |
| **State**       | The active scene + variant.                                                                                                                                        | `engine/scenes.State`                  |
| **Transition**  | Moving between two states. Holds the full resolved instructions, the changes to send, and the targets that are no longer instructed.                                | `engine/scenes.Transition`, `Plan`     |

### Why targets instead of devices

Scenes are prepared ahead of time and shared between sessions; hardware changes (a TV is swapped, a light is added, the game moves to another house). Keeping scenes device-agnostic means the only thing that changes is the room's bindings. A target can be bound to several devices (all lights in `ambient`), and a device can serve several targets.

A target name is shared across kinds: `table` can have a screen instruction _and_ a light instruction, and each is only sent to devices of the matching kind bound to `table`.

## Rules

### Resolving instructions

`Scene.Resolve(variantID)`:

1. Empty `variantID` → use `Scene.DefaultVariant`. If that is empty too, use base instructions only.
2. Start from the scene's base instructions.
3. For each kind, the variant's instruction for a target **replaces** the base instruction for that target (no field-level merging). Targets the variant does not mention keep their base instruction.
4. Unknown variant → `ErrVariantNotFound`.

Resolve never mutates the scene.

### Transitions

`scenes.Plan(current, currentVariant, next, nextVariant)` handles both _variant transitions_ (same scene) and _scene transitions_ — they are the same operation.

- `Changes` contains only instructions that are new or differ from the current state, so a variant that only changes the lights does not reload the battle map.
- `Released` lists targets the current state instructs but the next one does not. What to do with them (blank the screen, stop the music, leave as is) is **not decided yet** — `SceneService.Activate` currently ignores them.
- Fading / cross-fading is part of the instruction (`LightInstruction.Fade`, `MusicInstruction.CrossFade`), so each integration handles it natively.

### Dispatching

`application.SceneService.Activate(ctx, sceneID, variantID)`:

1. Load the scene (`SceneRepository`) and the room (`RoomProvider`).
2. `Plan` the transition from the active state.
3. Store the new active state.
4. For every changed instruction, find the devices bound to its target with the matching kind, and hand each to the driver registered for `device.Integration`.
5. Device errors (including `ErrNoDriver`) are joined and returned; other devices are still instructed.

## Outputs

### Screens — views

A `ScreenInstruction` selects a view; the client renders it. Screens are browser clients connected over the screen WebSocket (see [communication.md](./communication.md)).

| View        | Purpose                                                       | Typical `Asset`                    |
| ----------- | ------------------------------------------------------------- | ---------------------------------- |
| `battlemap` | Tactical grid with entities, fog of war, tokens               | `.dd2vtt` / `.dungeondraft_map`    |
| `scenery`   | Atmospheric image or video backdrop                           | image / video                      |
| `infoboard` | Handouts, initiative order, quest log, text                   | document / none                    |
| `custom`    | Client component chosen by `Component`, configured by `Params`| anything                           |
| `splash`    | Title card / loading screen                                   | image                              |
| `blank`     | Black screen                                                  | —                                  |

The existing `engine/screens` package (`ScreenHandler`, `ScreenManager`, `state.ScreenState`) is the transport for the `screen` integration: a screen driver will translate a `ScreenInstruction` plus the scene's entities into `ScreenState` / proto messages and broadcast it to the screen's listeners. The intended mapping is one `ScreenHandler` per screen device, so the WebSocket `screenId` becomes the device ID and routing by role or device is answered by the room's bindings.

### Lights

`LightInstruction`: `On`, `Color` (hex), `Brightness` (0–1), `Effect` (integration specific, e.g. `candle`), `Fade`. Planned integrations: Philips Hue, Matter, Home Assistant, custom.

### Music

`MusicInstruction`: `Action` (`play`/`pause`/`stop`), `Source` (integration specific URI, e.g. `spotify:playlist:...`), `Volume` (0–1), `Shuffle`, `Loop`, `CrossFade`. Planned integrations: Spotify, iTunes / Apple Music, custom.

### Custom devices

`CustomInstruction`: `Topic` + opaque JSON `Payload`, passed through untouched. Intended for MQTT first (fog machines, rumble packs, Arduino props), but any event bus can implement a driver.

## Code map

```ascii
server/
├── engine/
│   ├── devices/            Device, Kind, Room (target → device bindings), integration identifiers
│   ├── scenes/             Scene, Entity, Variant, Instructions, State, Transition, Plan
│   └── screens/            Screen WebSocket fan-out — the transport for the "screen" integration
└── application/
    └── scene_service.go    SceneService + ports: SceneRepository, RoomProvider, OutputDriver
```

## Status

**Implemented**: the domain model above, variant resolution, transition planning with change detection, `SceneService` dispatching to drivers via ports, unit tests for each.

**Not implemented yet**:

- Persistence: no `SceneRepository` / `RoomProvider` implementation, no DB models or migrations. Scenes are campaign-owned; rooms are global (or per server).
- `SceneService` is not wired in `components`.
- No drivers: screen (to `engine/screens`), Hue, Matter, Home Assistant, Spotify, iTunes, MQTT.
- No proto / ConnectRPC service for GM scene control (create/edit scenes, activate, switch variant).
- Screen routing by device: `WebsocketHandler` still takes a raw `screenId`.
- Handling of `Transition.Released`.
- Device registration / authentication (see [authentication.md](./authentication.md#device-authentication)).

## Extending

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
