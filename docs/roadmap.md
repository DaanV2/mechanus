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
- [x] User accounts and auth — see [Auth](#auth) below
- [x] Admin user list page (`/users`, `UserService.List`)
- [x] ConnectRPC services (users, login) and the screen WebSocket (`Ping`, `InitialSetupRequest`)
- [x] Database backends (SQLite, Postgres, MySQL, in-memory) with migrations
- [x] mDNS announcement of the server on the LAN
- [x] SvelteKit client shell with login / signup / profile and placeholder GM / player / device views
- [x] Scene model: config vs. state, variants, targets, rooms, transitions, `SceneService`, `scenestore`

### Auth

Done: per-login sessions (one JTI each) with an RS512 access token (1 h) and a `scope: "refresh"` token (7 days); refresh rotation and logout; client-side refresh and `exp` checks; `Unauthenticated` vs. `PermissionDenied` via `grpc.requireRole`; WebSocket auth by header, `mechanus.bearer.<jwt>` subprotocol or cookie, with connections closed on expiry / revocation; admin account creation, password reset and role management; self-service password change; open sign-up behind `auth.signup-enabled`; login lockout and timing-safe unknown users; signing key rotation; configurable lifetimes; cookie `Max-Age` / `Secure`; optional TLS. See [authentication.md](./architecture/authentication.md). Device auth is tracked under [M1](#m1--first-scene-on-the-table).

- [x] Invites (`InviteService`, `/users/invites`): admin generates a single use link / QR code with expiry, player picks their own credentials on `/users/signup?invite=<id>`
- [x] Admin UI to create accounts and manage roles (`/users`)
- [~] Screen access check (`ConnectionInfo.CanAccess`, `403` when denied, after the route policy): users open every screen, a paired device only its own. Per-screen ownership waits on the screen / room ownership model
- [x] Client reconnects the WebSocket with a refreshed token after the server closes it on expiry; a refresh that fails on the network no longer logs the user out
- [x] Honour `X-Forwarded-Proto` / `-Host` / `-For` behind a reverse proxy (`server.trust-proxy-headers`)
- [x] Tests: WebSocket 401 / 403, device ID / type from header and query, close on expiry / revocation
- [x] Playwright: logout, session survives token expiry via refresh, revoked session is logged out, WebSocket connects as the logged-in user, change password, invites, admin account / role management
- [ ] Move security policies into the message handlers: `DeviceService` off `grpc.PolicyInterceptor`, and split handlers that mix checks and work into an entry function and a private variant. Whether the screen WebSocket's `http.RequirePolicy` moves too — TBD

**Decisions**

- Security policies are enforced in the message handler, not in an interceptor. Preferably the entry function (`Create(ctx)`) handles the validations and a private variant (`create(ctx)`) does the actual creation of data, see [authentication.md](./architecture/authentication.md#where-security-checks-go)
- Accounts come from GM invites or admin creation; open sign-up only behind a config flag, off by default. Until a GM role exists, invites need the admin role
- Invites are single use, expire after a chosen duration (default 7 days, at most 30) and give the `user` role; the invite ID is the secret in the link
- Tokens stay in JS-readable `SameSite=Lax` cookies (not `HttpOnly` / in-memory); same-origin cookies also authenticate the WebSocket upgrade
- Plain HTTP by default, optional TLS via config and reverse-proxy support. Cookies get `Secure` only over HTTPS. Screens need a wake-lock fallback, since the Screen Wake Lock API requires a secure context

---

## M1 — First scene on the table

**Goal**: the GM picks a scene in the GM view, and a TV registered as a device shows that scene's battle map (or scenery). Switching variant updates the TV. Everything else can be rough.

**Server**

- [x] `SceneService` + `scenestore` wired into `components` (single room in `rooms/room.dat`, `Exit` on `BeforeShutdown`)
- [x] Screen state proto (`ScreenState`: view, overlays, layers, entities, grid, camera) sent on initial setup and via `ScreenHandler.SetState`; entity / camera delta messages defined
- [x] Screen driver (`infrastructure/outputs/screen`): view, overlays, grid settings, layers and visible entities are pushed; asset paths resolve through `infrastructure/assets`
- [x] Screen WebSocket connections routed by device ID
- [x] Device identity from `X-Device-ID` / `X-Device-Type` (query fallback `deviceId` / `deviceType`) stored in `screens.ConnectionInfo`, logs and the connection's trace span (`device.id`, `device.type`, `screen.id`)
- [x] Scene control API (`scenes/v1`): list, activate, switch variant, get active state — any logged-in user until a GM role exists
- [x] Scene asset serving with content-hashed URLs (`/assets/scenes/<scene id>/...`, `/assets/shared/...`)
- [x] Per-route auth policies (`anonymous`, `device`, `user`, `role:<role>`): auth optional at connect, enforced by the route; invalid credentials rejected instead of treated as anonymous — on the screen WebSocket and `DeviceService`, see [authentication.md](./architecture/authentication.md#device-authentication)
- [x] Device registration: stable ID, bindable to a target (`DeviceService.Bind`, merged into the room's bindings)
- [x] Device pairing (`DeviceService.Pair`) issuing a revocable device token — the screen WebSocket no longer accepts connections without a credential
- [x] Battle map format v0: grid settings + ordered image and entity layers, with instance groups (`*.map.json`, see [battle-maps.md](./configuration/battle-maps.md))
- [x] Example battle map scene for development (`docs/examples/battlemap`, mounted by the compose files)

**Client**

- [x] Device / TV view (`/views/devices/`): fullscreen Pixi.js renderer driven by screen state, auto-reconnect, `?device=` / `?type=` overrides
- [x] Views: `battlemap` (layers + grid), `scenery` (image or video), `splash`, `blank`
- [x] GM view v0 (`/views/game-master/`): list scenes, activate one, switch variant
- [x] Layer system in `src/lib/2d/layers` (image layers, grid, entities, overlays) and render loop; instance groups via a Pixi `ParticleContainer` per group
- [x] Device pairing page (`/devices/pair`): shows the 4-digit code full screen, renews it, stores the device token; the device view sends a device without credentials here, and back here once revoked
- [x] Devices admin page (`/devices`): enter a pairing code + name, list paired / connected devices, bind to a target, revoke

**Decisions**

- A battle map is grid settings + an ordered list of typed layers. v0 has image and entity layers (including instance groups: one asset at many positions). Walls, lights, sound, ... are added as layer kinds with the feature that needs them
- A battle map is its own file in the scene's folder (`maps/cave.map.json`), named by the screen instruction's `asset`, so variants can swap maps. The server reads it and sends screens only what they may see, so map files are not served under `/assets/`
- Each scene is a folder with its config and assets; shared folders hold reused assets. Paths are scene-relative (`maps/cave.webp`) or prefixed (`shared:/monsters/goblin.png`)
- Pairing: the device shows a 4-digit code, an admin types it in. Codes are streamed by `DeviceService.Pair` (`anonymous` route), valid only while the stream is open (max 5 min, auto-renewed), confirm attempts rate-limited: 5 wrong codes within a minute replace every open code and block that admin for the rest of the minute
- A device's target binding is state of the registration (database), not authored room config: it is merged into the room when the room is read, the room file stays read only. It applies from the next instruction of a scene, and a revoked device loses it
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

- [~] Entity tokens on the battle map: tokens are drawn at the position from the map file; still to do: positions stored in scene state
- [ ] Screen WebSocket messages for pan, zoom, token moves
- [ ] GM controls: reveal / hide / spawn / remove entities
- [ ] Token control: `controlled_by` users per entity, GM assigns / shares control, live movement lock
- [ ] Fog of war and reveal, stored in scene state
- [~] Overlays: `entities` lists names on battle maps; still to do: portraits, the same on other views, and the `initiative` tracker
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

- [x] Driver framework: integration config sets (credentials, bridge / broker addresses), health / connection status per device (`DeviceMonitor`), per-instruction timeout, devices instructed concurrently — see [scenes.md](./architecture/scenes.md#driver-framework)
- [ ] Show device status to the GM: API and GM view on top of `DeviceMonitor.Statuses`
- [ ] Lights: Philips Hue
- [ ] Lights: Home Assistant
- [ ] Lights: Matter (native Go controller: commissioning, fabric management)
- [ ] Music: Spotify
- [ ] Music: iTunes / Apple Music
- [x] Custom devices: MQTT (`infrastructure/outputs/mqtt`), set up in [configuration/integrations.md](./configuration/integrations.md)
- [ ] Third-party event API: external software subscribes to scene events and sends events back
- [ ] Handle released targets (`Transition.Released`) — reset, stop, or keep
- [ ] GM live controls: adjust lights / volume mid-scene (`SceneService.Adjust`)

**Decisions**

- No fixed integration order; the driver framework comes first since everything builds on it
- Home Assistant, Philips Hue and Matter are three separate native integrations. Matter means writing a Go controller, with no mature library to lean on
- MQTT: scenes hold an opaque payload per target; the topic comes from the device config (default `mechanus/devices/<device-id>/set`). Published retained, wrapped in `{scene, variant, target, payload}`. Scene events go to `mechanus/events/scene` (not retained), availability to `mechanus/status` as `online` / `offline` (retained, last will)
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
- [x] Packaging: single binary for Linux, Windows and macOS (amd64 + arm64), multi-arch Docker image, all tested in CI. GoReleaser (`.goreleaser.yaml`) builds the archives (binary + `web/`) on release, and PR CI builds all of them in one job and starts the linux/amd64 one; the image is built for both arches on every PR. See [installation.md](./installation.md)
- [ ] Serve the web client without `--server.static.folder`: embed it in the binary, or default to a `web/` folder next to the executable. Which one — TBD
- [x] About page content and licenses (`client/src/routes/about`): third-party licences are generated at build time (`server/scripts/licenses` for Go modules, `client/scripts/vite-plugin-licenses.ts` for bundled npm packages)
- [x] macOS configuration paths documented (`docs/configuration/config-file.md`)
- [ ] User-facing documentation: setting up a room, writing scenes

**Decisions**

- Linux, Windows, macOS (amd64 + arm64) and a multi-arch Docker image are all fully supported — covers Raspberry Pi, NAS, mini PC and laptops. SQLite is pure Go, so cross-compiling needs no cgo
- Release archives ship the web client as a `web/` folder next to the binary, the way the server serves it today (`server.static.folder`)

---

## Parking lot

Ideas that don't belong to a milestone yet.

- Keep screen state in order per connection: `ScreenHandler.Broadcast` writes each message in its own goroutine, so two quick changes (a GM switching variant twice) can arrive out of order. Each `ScreenState` replaces everything on the client, so the TV would stay on the older one until the next change or reconnect. Initial setup replies (`HandleMessages`) can overtake a newer state the same way. Seen in the code, not yet in practice
