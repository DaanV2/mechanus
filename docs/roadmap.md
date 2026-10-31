# Roadmap

High-level overview of what needs to be built to get Mechanus from "foundation" to "the whole game room runs from the browser". Milestones are ordered by dependency, not by date. Each one has a **goal** (how we know it's done), the **features** it needs, and **unknowns** that still have to be decided.

Conventions:

- `[ ]` not started · `[~]` partially there · `[x]` done
- `TBD` marks an open decision or a detail still to be filled in. Search for `TBD` to find them.
- Details of the scene model live in [architecture/scenes.md](./architecture/scenes.md); this file only tracks _what_ to build, not _how_.

## Overview

```ascii
 M0 Foundation ──► M1 First scene on the table ──► M3 Live play on the battle map ──► M5 Sessions & campaigns
                         │                                    ▲
                         ├──► M2 Scene authoring ─────────────┘
                         │
                         └──► M4 The room comes alive (lights, music, custom devices)
                                                                           M6 Polish & release (ongoing)
```

| Milestone                                              | Goal in one line                                                        |
| ------------------------------------------------------ | ----------------------------------------------------------------------- |
| [M0 Foundation](#m0--foundation)                       | Server, auth, transport, and the scene model exist                      |
| [M1 First scene on the table](#m1--first-scene-on-the-table) | The GM activates a scene and the table TV shows its battle map    |
| [M2 Scene authoring](#m2--scene-authoring)             | Scenes and rooms can be made without hand-writing files                 |
| [M3 Live play on the battle map](#m3--live-play-on-the-battle-map) | Tokens, fog, initiative and player views work during a session |
| [M4 The room comes alive](#m4--the-room-comes-alive)   | Lights, music and custom devices follow the scene                       |
| [M5 Sessions & campaigns](#m5--sessions--campaigns)    | Campaigns, players and sessions are managed in the app                  |
| [M6 Polish & release](#m6--polish--release)            | Installable, discoverable, documented                                   |

---

## M0 — Foundation

**Goal**: the building blocks every later milestone relies on.

- [x] Go server with Clean Architecture layers, lifecycle, config, logging, OpenTelemetry
- [x] User accounts, login, JWT / JTI auth, roles
- [x] Admin user list page (`/admin/users`, `UserService.List`)
- [x] ConnectRPC services (users, login) and the screen WebSocket (`Ping`, `InitialSetupRequest`)
- [x] Database backends (SQLite, Postgres, MySQL, in-memory) with migrations
- [x] mDNS announcement of the server on the LAN
- [x] SvelteKit client shell with login / signup / profile and placeholder GM / player / device views
- [~] Scene model: config vs. state, variants, targets, rooms, transitions, `SceneService`, `scenestore` (code + tests, not wired)

**Unknowns**

- TBD

---

## M1 — First scene on the table

**Goal**: the GM picks a scene in the GM view, and a TV registered as a device shows that scene's battle map (or scenery). Switching variant updates the TV. Everything else can be rough.

**Server**

- [ ] Wire `SceneService` + `scenestore` into `components` (room ID config flag, `Exit` on `BeforeShutdown`)
- [ ] Screen driver: turn a `ScreenInstruction` + scene entities into screen state and push it over the WebSocket
- [ ] Route screen WebSocket connections by device ID (one screen per screen device)
- [ ] Device registration: a device identifies itself, gets a stable ID, and can be bound to a target
- [ ] Device authentication (API key / pairing) — currently any non-JWT connection is accepted
- [ ] Scene control API (`scenes/v1` proto): list scenes, activate, switch variant, get active state
- [ ] Asset serving for scene assets (maps, images, video)
- [ ] Battle map format v0 — just enough to render a background and grid

**Client**

- [ ] Device / TV view: fullscreen Pixi.js renderer driven by server screen state
- [ ] Layer system in `src/lib/2d` (background, grid, entities, overlays) and the render loop
- [ ] Views: `battlemap` (background + grid), `scenery`, `splash`, `blank`
- [ ] GM view v0: list scenes, activate one, switch variant
- [ ] Device setup page: show device ID / pairing state

**Unknowns**

- Battle map format: what goes in v0 (grid, walls, lighting, fog?) and how it is stored — TBD
- Where assets live (file storage folder per campaign? content-addressed?) and how they are referenced from configs — TBD
- Device pairing UX: code on the TV screen, QR code, or GM approves new devices? — TBD
- Which room is active when there's more than one — TBD
- TBD

---

## M2 — Scene authoring

**Goal**: a GM can create and edit scenes, variants and the room setup from the app, and import existing maps, without touching JSON files.

- [ ] Config schema with versioning (proto `scenes/v1` as file + API schema, or a `version` field with migrations)
- [ ] Scene editor: entities, base instructions per target, variants
- [ ] Room editor: devices, target bindings
- [ ] Asset upload and management
- [ ] Map import: Universal VTT (`.dd2vtt`) and Dungeondraft into the Mechanus battle map format (`mechanus maps import` is a stub today)
- [ ] Preview a scene / variant without activating it on the room

**Unknowns**

- In-app editor only, or also support editing config files directly (and reload on change)? — TBD
- File format for configs: JSON, YAML, proto JSON? — TBD
- Are scenes shareable between campaigns (templates / library)? — TBD
- TBD

---

## M3 — Live play on the battle map

**Goal**: a full combat encounter can be run on the table: tokens move, fog is revealed, initiative is tracked, and players see only what they should.

- [ ] Entity tokens on the battle map, positions stored in scene state
- [ ] Screen WebSocket messages for pan, zoom, token moves
- [ ] GM controls: reveal / hide / spawn / remove entities
- [ ] Fog of war and reveal, stored in scene state
- [ ] Overlays: `initiative` tracker, `entities` portraits
- [ ] Info board view (handouts, initiative order, quest log)
- [ ] Player view with limited visibility
- [ ] Reset a scene's state back to its config

**Unknowns**

- Player views: a target per player, or one role-based view derived from the battle map? — TBD
- Who may move tokens: GM only, or players for their own character? — TBD
- How much rules knowledge (HP, conditions, initiative rolls) belongs in Mechanus vs. stays system-agnostic — TBD
- Touch input on a table-mounted TV — TBD
- TBD

---

## M4 — The room comes alive

**Goal**: activating or switching a scene also sets the lights, starts the right music, and triggers custom devices.

- [ ] Driver framework: integration config sets (credentials, bridge / broker addresses), health / connection status per device
- [ ] Lights: Philips Hue
- [ ] Lights: Home Assistant
- [ ] Lights: Matter
- [ ] Music: Spotify
- [ ] Music: iTunes / Apple Music
- [ ] Custom devices: MQTT
- [ ] Handle released targets (`Transition.Released`) — reset, stop, or keep
- [ ] GM live controls: adjust lights / volume mid-scene (`SceneService.Adjust`)

**Unknowns**

- Which integration first — TBD
- Auth flows that expect a public callback URL (Spotify OAuth) on a LAN-only server — TBD
- Is Home Assistant enough to cover Hue / Matter (one integration instead of three)? — TBD
- Custom device payload conventions (topic naming, schemas) — TBD
- TBD

---

## M5 — Sessions & campaigns

**Goal**: campaigns, players and play sessions are managed in the app, and a session can be stopped and resumed exactly where it left off.

- [ ] Campaign management UI: create, edit, delete campaigns; invite / manage players
- [ ] Characters linked to scene entities (`Entity.CharacterID`)
- [ ] Remember the active scene across server restarts
- [ ] Session concept: start / end a session, which scenes were played
- [ ] Logout flow in the client (`src/lib/handlers/user.ts`)

**Unknowns**

- Is a "session" a first-class object, or just the active campaign + scene? — TBD
- Should scene state be kept per session (history) or only the latest? — TBD
- TBD

---

## M6 — Polish & release

**Goal**: someone else can install Mechanus on a home network and use it.

- [ ] Client discovers the server via mDNS (server side already announces)
- [ ] Packaging: single binary releases, Docker image, setup docs
- [ ] About page content and licenses (`client/src/routes/about`)
- [ ] macOS configuration paths documented (`docs/configuration/config-file.md`)
- [ ] User-facing documentation: setting up a room, writing scenes

**Unknowns**

- Target platforms (Raspberry Pi? NAS? desktop?) — TBD
- TBD

---

## Parking lot

Ideas that don't belong to a milestone yet.

- TBD
