# Roadmap

What still has to be built to get Mechanus from "foundation" to "the whole game room runs from the browser". Milestones are ordered by dependency, not by date. Each one has a **goal** (how we know it's done), the **features** it needs, the **decisions** already made, and the **open questions** still to settle.

Conventions:

- `[ ]` not started · `[~]` partially there · `[x]` done
- `TBD` marks an open decision. Search for `TBD` to find them.
- This file tracks _what_ to build, not _how_. The scene model is described in [architecture/scenes.md](./architecture/scenes.md).

## Overview

```ascii
 M0 Foundation ──► M1 First scene ──► M3 Live play ──► M5 Sessions & campaigns
                        │                 ▲
                        ├──► M2 Authoring ┘
                        │
                        └──► M4 The room comes alive

 M6 Polish & release runs alongside everything
```

| Milestone                                                          | Goal in one line                                                 |
| ------------------------------------------------------------------ | ---------------------------------------------------------------- |
| [M0 Foundation](#m0--foundation)                                   | Server, auth, transport, and the scene model exist               |
| [M1 First scene on the table](#m1--first-scene-on-the-table)       | The GM activates a scene and the table TV shows its battle map   |
| [M2 Scene authoring](#m2--scene-authoring)                         | Scenes and rooms can be made without hand-writing files          |
| [M3 Live play on the battle map](#m3--live-play-on-the-battle-map) | Tokens, fog, initiative and player views work during a session   |
| [M4 The room comes alive](#m4--the-room-comes-alive)               | Lights, music and custom devices follow the scene                |
| [M5 Sessions & campaigns](#m5--sessions--campaigns)                | Campaigns, players and sessions are managed in the app           |
| [M6 Polish & release](#m6--polish--release)                        | Installable, discoverable, documented                            |

---

## M0 — Foundation

**Goal**: the building blocks every later milestone relies on.

- [x] Go server with Clean Architecture layers, lifecycle, config, logging, OpenTelemetry
- [~] User accounts and auth — see [Auth](#auth) below for what is still missing
- [x] Admin user list page (`/users`, `UserService.List`)
- [x] ConnectRPC services (users, login) and the screen WebSocket (`Ping`, `InitialSetupRequest`)
- [x] Database backends (SQLite, Postgres, MySQL, in-memory) with migrations
- [x] mDNS announcement of the server on the LAN
- [x] SvelteKit client shell with login / signup / profile and placeholder GM / player / device views
- [x] Scene model: config vs. state, variants, targets, rooms, transitions, `SceneService`, `scenestore`

### Auth

Done: per-login sessions (one JTI each) with an RS512 access token (1 h) and a `scope: "refresh"` token (7 days); refresh rotation and logout; client-side refresh and `exp` checks; `Unauthenticated` vs. `PermissionDenied` via `grpc.requireRole`; WebSocket auth by header, `mechanus.bearer.<jwt>` subprotocol or cookie, with connections closed on expiry / revocation; admin account creation, password reset and role management; self-service password change; open sign-up behind `auth.signup-enabled`; login lockout and timing-safe unknown users; signing key rotation; configurable lifetimes; cookie `Max-Age` / `Secure`; optional TLS. See [authentication.md](./architecture/authentication.md). Device auth is tracked under [M1](#m1--first-scene-on-the-table).

- [ ] Invites: GM generates a link / QR code with expiry, player picks their own credentials
- [ ] Admin UI to create accounts and manage roles
- [ ] Screen access check (`// TODO: check that screenHandler is allowed to be access by this requester`) — needs the screen / room ownership model
- [ ] Client reconnects the WebSocket with a refreshed token after the server closes it on expiry
- [ ] Honour `X-Forwarded-Proto` behind a reverse proxy
- [ ] Tests: WebSocket route policies (401 / 403), device ID / type from header and query, close on expiry / revocation (needs a way to register a screen in tests)
- [ ] Playwright: logout, session survives token expiry via refresh, revoked session is logged out, WebSocket connects as the logged-in user, change password

**Decisions**

- Accounts come from GM invites or admin creation; open sign-up only behind a config flag, off by default
- Tokens stay in JS-readable `SameSite=Lax` cookies (not `HttpOnly` / in-memory); same-origin cookies also authenticate the WebSocket upgrade
- Plain HTTP by default, optional TLS via config and reverse-proxy support. Cookies get `Secure` only over HTTPS. Screens need a wake-lock fallback, since the Screen Wake Lock API requires a secure context

---

## M1 — First scene on the table

**Goal**: the GM picks a scene in the GM view, and a TV registered as a device shows that scene's battle map (or scenery). Switching variant updates the TV. Everything else can be rough.

**Server**

- [x] `SceneService` + `scenestore` wired into `components` (single room in `rooms/room.dat`, `Exit` on `BeforeShutdown`)
- [x] Screen state proto (`ScreenState`: view, overlays, layers, entities, grid, camera) sent on initial setup and via `ScreenHandler.SetState`; entity / camera delta messages defined
- [~] Screen driver (`infrastructure/outputs/screen`): view, overlays and asset URL are pushed; entities, grid and layers wait on the battle map format
- [x] Screen WebSocket connections routed by device ID
- [~] Device identity from `X-Device-ID` / `X-Device-Type` (query fallback `deviceId` / `deviceType`) stored in `screens.ConnectionInfo` and logs; not yet on trace spans
- [x] Scene control API (`scenes/v1`): list, activate, switch variant, get active state — any logged-in user until a GM role exists
- [x] Scene asset serving with content-hashed URLs (`/assets/scenes/<scene id>/...`, `/assets/shared/...`)
- [ ] Per-route auth policies (`anonymous`, `device`, `user`, `role:<role>`): auth optional at connect, enforced by the route; invalid credentials rejected instead of treated as anonymous — see [authentication.md](./architecture/authentication.md#device-authentication)
- [ ] Device registration: stable ID, bindable to a target
- [ ] Device pairing (`DeviceService.Pair`) issuing a revocable device token — today any non-JWT connection is accepted
- [ ] Battle map format v0: grid settings + ordered image and entity layers, with instance groups

**Client**

- [x] Device / TV view (`/views/devices/`): fullscreen Pixi.js renderer driven by screen state, auto-reconnect, `?device=` / `?type=` overrides
- [x] Views: `battlemap` (background + grid), `scenery` (image or video), `splash`, `blank`
- [x] GM view v0 (`/views/game-master/`): list scenes, activate one, switch variant
- [ ] Layer system in `src/lib/2d` (image layers, grid, entities, overlays) and render loop; instance groups via Pixi instancing
- [~] Device pairing page (`/devices/pair`): device ID generation done (`$lib/devices/identity.ts`); still to do: show the 4-digit code full screen, renew it, store the device token
- [ ] Devices admin page (`/devices`): enter a pairing code + name, list paired / connected devices, revoke

**Decisions**

- A battle map is grid settings + an ordered list of typed layers. v0 has image and entity layers (including instance groups: one asset at many positions). Walls, lights, sound, ... are added as layer kinds with the feature that needs them
- Each scene is a folder with its config and assets; shared folders hold reused assets. Paths are scene-relative (`maps/cave.webp`) or prefixed (`shared:/monsters/goblin.png`)
- Pairing: the device shows a 4-digit code, an admin types it in. Codes are streamed by `DeviceService.Pair` (`anonymous` route), valid only while the stream is open (max 5 min, auto-renewed), confirm attempts rate-limited
- Device tokens: JWT backed by a JTI (`Authorization: Device <token>`), 8 h access + 30-day refresh; a device unused for 30 days has to pair again
- `X-Device-Type` values: `screen`, `table`, `tablet`, `phone`, `desktop`, `other`; unknown values become `other`. Metadata only, not used for authorization
- One running instance serves one game table, so there is exactly one room — no room selection

---

## M2 — Scene authoring

**Goal**: a GM can create and edit scenes, variants and the room setup from the app, and import existing maps, without touching JSON files.

- [ ] Config files in YAML or JSON with a `version` field and migrations on load; translated to `scenes/v1` proto for the API / screens
- [ ] Scene editor: entities, base instructions per target, variants
- [ ] Room editor: devices, target bindings
- [ ] Asset upload and management
- [ ] Map import: Universal VTT (`.dd2vtt`) and Dungeondraft (`mechanus maps import` is a stub today)
- [ ] Preview a scene / variant without activating it
- [ ] Scene state keyed by campaign + scene
- [ ] Live reload of hand-edited files: watch folders, reload, re-instruct devices if the active scene changed, report invalid files in the GM view and keep the last good version

**Decisions**

- Both an in-app editor (primary) and hand-editable files, reloaded on change
- YAML and JSON both supported, picked by extension. The proto is the wire contract, not the file schema
- Scenes are campaign-independent; their state is saved per campaign + scene

---

## M3 — Live play on the battle map

**Goal**: a full combat encounter can be run on the table: tokens move, fog is revealed, initiative is tracked, and players see only what they should.

- [ ] Entity tokens on the battle map, positions stored in scene state
- [ ] Screen WebSocket messages for pan, zoom, token moves
- [ ] GM controls: reveal / hide / spawn / remove entities
- [ ] Token control: `controlled_by` users per entity, GM assigns / shares control, live movement lock
- [ ] Fog of war and reveal, stored in scene state
- [ ] Overlays: `initiative` tracker, `entities` portraits
- [ ] Game primitives: resource bars, condition tags with round counters, initiative order with turn / round tracking, free-form entity stats
- [ ] Info board view (handouts, initiative order, quest log)
- [ ] Player view with limited visibility, overridable by a scene's `players` target
- [ ] Reset a scene's state back to its config
- [ ] Touch input on screens: per-device `touch` mode, live GM touch lock

**Decisions**

- Player views are derived server-side per user from battle map state (hidden / fogged entities never sent, camera on their character). A scene can instruct a `players` target to override this for special moments
- Each entity has a `controlled_by` list (a character defaults to its player); the GM can move everything, assign or share control, and freeze player movement
- No rules engine and no dice — only generic primitives: campaign-defined resource bars (current / max, visible to everyone / owner / GM), user-defined condition tags with icons and optional round counter, hand-entered initiative values, key/value stats
- Touch is a per-device room setting: `none`, `view` (pan, zoom, ping, measure; the default) or `control` (moves tokens as a configured role or user)

---

## M4 — The room comes alive

**Goal**: activating or switching a scene also sets the lights, starts the right music, and triggers custom devices.

- [ ] Driver framework: integration config sets (credentials, bridge / broker addresses), health / connection status per device
- [ ] Lights: Philips Hue
- [ ] Lights: Home Assistant
- [ ] Lights: Matter (native Go controller: commissioning, fabric management)
- [ ] Music: Spotify
- [ ] Music: iTunes / Apple Music
- [ ] Custom devices: MQTT
- [ ] Third-party event API: external software subscribes to scene events and sends events back
- [ ] Handle released targets (`Transition.Released`) — reset, stop, or keep
- [ ] GM live controls: adjust lights / volume mid-scene (`SceneService.Adjust`)

**Decisions**

- No fixed integration order; the driver framework comes first since everything builds on it
- Home Assistant, Philips Hue and Matter are three separate native integrations. Matter means writing a Go controller, with no mature library to lean on
- MQTT: scenes hold an opaque payload per target; the topic comes from the device config (default `mechanus/devices/<device-id>/set`). Published retained, wrapped in `{scene, variant, target, payload}`. Scene events go to `mechanus/events/scene`, availability to `mechanus/status` (last will)
- Third-party event API: JSON WebSocket only, same event schema as MQTT, authenticated with admin-created, revocable, scoped API tokens (JWT + JTI)

**Open questions**

- Auth flows that need a public callback URL (Spotify OAuth) on a LAN-only server. Options: project-hosted static HTTPS relay page (PKCE, user's own client ID), loopback redirect + paste the URL back, delegate to Home Assistant, or require TLS — TBD
- Event API scopes: candidates `events:read` (subscribe), `events:write` (custom events in the GM view), `scenes:control` (activate / switch variant / adjust) — TBD

---

## M5 — Sessions & campaigns

**Goal**: campaigns, players and play sessions are managed in the app, and a campaign's session continues exactly where it left off — across server restarts and campaign switches.

- [x] Campaign overview and detail pages: users see their campaigns, admins see all; users can create one
- [ ] Campaign management UI: edit, delete, invite / manage players
- [ ] Characters linked to scene entities (`Entity.CharacterID`)
- [ ] Remember the active scene across server restarts
- [ ] Session object: one continuous session per campaign holding the active scene and play state

**Decisions**

- A session is the never-ending game of a campaign — no start / end lifecycle; switching campaign switches session, returning resumes it
- Scene state keeps only the latest version per campaign + scene, overwritten on exit; no history

---

## M6 — Polish & release

**Goal**: someone else can install Mechanus on a home network and use it.

- [ ] Client discovers the server via mDNS (the server already announces)
- [ ] Packaging: single binary for Linux, Windows and macOS (amd64 + arm64), multi-arch Docker image, all tested in CI
- [ ] About page content and licenses (`client/src/routes/about`)
- [ ] macOS configuration paths documented (`docs/configuration/config-file.md`)
- [ ] User-facing documentation: setting up a room, writing scenes

**Decisions**

- Linux, Windows, macOS (amd64 + arm64) and a multi-arch Docker image are all fully supported — covers Raspberry Pi, NAS, mini PC and laptops. SQLite is pure Go, so cross-compiling needs no cgo

---

## Parking lot

Ideas that don't belong to a milestone yet.

- _Nothing yet._
