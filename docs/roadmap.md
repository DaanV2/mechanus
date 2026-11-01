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
- [~] User accounts, login, JWT / JTI auth, roles — works on the happy path, see [Auth & tokens](#auth--tokens) for what's broken or missing
- [x] Admin user list page (`/users`, `UserService.List`)
- [x] ConnectRPC services (users, login) and the screen WebSocket (`Ping`, `InitialSetupRequest`)
- [x] Database backends (SQLite, Postgres, MySQL, in-memory) with migrations
- [x] mDNS announcement of the server on the LAN
- [x] SvelteKit client shell with login / signup / profile and placeholder GM / player / device views
- [x] Scene model: config vs. state, variants, targets, rooms, transitions, `SceneService`, `scenestore` (wired in `components` as part of M1)

### Auth & tokens

What exists: login (`LoginService.Login`) issues an RS512 JWT (1 h, `grant` / `scope` claims) tied to a per-user `JTI` row; the ConnectRPC `Authenticator` interceptor validates it and puts the claims on the context; `UserService` checks roles; admins can list users and reset passwords. `JTIService.Revoke` exists but nothing calls it. Device auth is tracked under [M1](#m1--first-scene-on-the-table).

**Bugs**

- [x] `grpc.Authenticator.validateAndInject` panicked (nil pointer) on any malformed, expired or badly signed bearer token — now treated as an anonymous request
- [x] Screen WebSocket never authenticated browser users (browsers can't set `Authorization` on `new WebSocket(...)`) — the client now passes its token as a `mechanus.bearer.<jwt>` subprotocol and the server also reads the `access-token` cookie (`authentication.BearerFromHeaders`)
- [x] Cookie fallback in `grpc.getJwtValue` never matched the URL-encoded `Bearer%20<jwt>` cookie — replaced by `authentication.BearerFromHeaders`, which unescapes it
- [ ] `Refresh` accepts any `scope: "*"` access token and reuses the same JTI, so a leaked token can be renewed forever; no `scope: "refresh"` token is ever issued (the client's `refresh-token` cookie key is unused).
- [ ] Admin password reset (`ResetPassword`) does not revoke the user's JTI — existing sessions stay valid.
- [ ] Tokens leak into logs: `JWTService.validate` logs the raw JWT at debug level, and the client `Cookie.get` / `Cookie.set` `console.log` the access token.
- [ ] Invalid / expired tokens are silently treated as anonymous; handlers then answer `PermissionDenied`, so the client can't tell "log in again" (`Unauthenticated`) from "not allowed".
- [ ] `KeyManager.GetSigningKey` picks whichever key the map yields first once more than one key exists — signing key is effectively random.
- [ ] `authentication.md` claims a 5-minute expiry leeway; `jwt.ParseWithClaims` checks `exp` with zero leeway before the leeway validator runs, so the leeway never applies.

**Missing**

- [ ] Account creation: invites (GM generates a link / QR code with expiry, player picks their own credentials), admin-created accounts, and an open sign-up config flag (default off) gating the unauthenticated `UserService.Create`
- [ ] Logout: server endpoint that revokes the JTI, plus the client flow (`UserHandler.logout` is a `TODO`)
- [ ] Proper refresh flow: login returns an access token + a `scope: "refresh"` token; the client refreshes before expiry and on `Unauthenticated`
- [ ] Client checks `exp` — `userHandler.current` reports logged-in for an expired token, and the cookie has no `Max-Age`
- [x] WebSocket auth for users: `Authorization` header, `mechanus.bearer.<jwt>` subprotocol, or `access-token` cookie
- [ ] The `// TODO: check that screenHandler is allowed to be access by this requester` screen access check
- [ ] Long-lived WebSocket connections are never re-checked when the token expires or the JTI is revoked
- [ ] Per-session JTIs (so one device can be signed out) instead of one shared JTI per user; cleanup of revoked / old JTI rows
- [ ] Self-service password change (revokes other sessions); admin role management API (role changes only reach the token on refresh)
- [ ] Login hardening: rate limiting / lockout, and constant-time behaviour when the username doesn't exist (today an unknown user returns before hashing, which leaks which usernames exist)
- [ ] Signing key rotation (new key on a schedule, retire old keys after the max token lifetime)
- [ ] Configurable token lifetime (`JWTOptions.TokenDuration` is hard-coded to 1 h) via a config set
- [ ] Cookie hardening: the token stays in a JS-readable `SameSite=Lax` cookie; add `Max-Age` matching the token expiry and set `Secure` when the request arrived over HTTPS
- [ ] Optional TLS: `SERVER_TLS_CERT` / `SERVER_TLS_KEY` config, and honour `X-Forwarded-Proto` behind a reverse proxy

**Tests**

- [x] `grpc.Authenticator`: no header, valid, malformed, expired, tampered, revoked and refresh-scope tokens, and the URL-encoded cookie fallback
- [ ] `LoginServiceHandler`: login success, wrong password, unknown user, empty fields; refresh with valid / revoked / refresh-scope / non-refresh-scope tokens
- [ ] `UserServiceHandler` authorization: `Get` (self / admin / other / anonymous), `List` and `ResetPassword` (admin vs. non-admin), reset revokes sessions
- [ ] `JWTService` negative cases: expired, not-yet-valid, wrong issuer / audience, unknown or missing `kid`, tampered signature, JTI belonging to another user
- [ ] `KeyData` PEM marshal / unmarshal round trip; `KeyManager` loads keys on `AfterInitialize`, reuses the signing key, errors on an unknown `kid`
- [~] WebSocket `authenticate`: user token accepted and bad token rejected (header, subprotocol and cookie), refresh-scope token rejected are covered; still missing: route policies return 401 / 403, device ID / type read from header and query fallback
- [ ] Playwright: logout, session survives token expiry via refresh, revoked session is logged out, WebSocket connects as the logged-in user

**Unknowns**

- ~~Is open sign-up intended?~~ Decided: accounts come from GM invites (link / QR) or admin creation; open sign-up only behind a config flag, off by default
- ~~Where tokens live on the client?~~ Decided: keep the JS-readable cookie; fix the encoding, logging and `Max-Age` bugs rather than moving to `HttpOnly` / in-memory storage. Same-origin cookies are still sent on the WebSocket upgrade, so the server can read the cookie for WebSocket auth
- ~~HTTPS on a LAN server?~~ Decided: plain HTTP by default; optional TLS via config (cert / key paths) and reverse-proxy support (`X-Forwarded-Proto`). Cookies get `Secure` only when the request arrived over HTTPS. Screens need a wake-lock fallback since the Screen Wake Lock API requires a secure context

---

## M1 — First scene on the table

**Goal**: the GM picks a scene in the GM view, and a TV registered as a device shows that scene's battle map (or scenery). Switching variant updates the TV. Everything else can be rough.

**Server**

- [x] Wire `SceneService` + `scenestore` into `components` (the instance's single room, `Exit` on `BeforeShutdown`) — room file is `rooms/room.dat`; a missing room is an empty room
- [x] Screen state proto: `ScreenState` (view, overlays, layers, entities, grid, camera) is sent on initial setup and via `ScreenHandler.SetState`; entity / camera delta messages defined
- [~] Screen driver: turn a `ScreenInstruction` + scene entities into screen state and push it over the WebSocket — view, overlays and asset URL are pushed (`infrastructure/outputs/screen`); entities, grid settings and layers wait on the battle map format
- [x] Route screen WebSocket connections by device ID (one screen per screen device, created on first connection or instruction)
- [ ] Per-route auth policies (`anonymous`, `device`, `user`, `role:<role>`): authentication is optional at connect, enforced by the route; invalid credentials are rejected instead of treated as anonymous — see [authentication.md](./architecture/authentication.md#device-authentication)
- [x] Device identity: read `X-Device-ID` / `X-Device-Type` (query fallback `deviceId` / `deviceType` for browsers) and store them in the connection's session info (`screens.ConnectionInfo`, logger / trace attributes) — stored and logged; not yet added as trace span attributes
- [ ] Device registration: a device identifies itself, gets a stable ID, and can be bound to a target
- [ ] Device pairing: `DeviceService.Pair` server stream pushes a 4-digit single-use code (valid while the stream is open, max 5 min, auto-renewed), confirmed by an admin, rate-limited; issues a revocable device JWT backed by a JTI (`Authorization: Device <token>`, 8 h) plus a device refresh token (30 days) the device uses to renew itself — currently any non-JWT connection is accepted
- [x] Scene control API (`scenes/v1` proto): list scenes, activate, switch variant, get active state — any logged-in user may control scenes until a GM role exists
- [x] Asset serving for scene assets (maps, images, video): resolve scene-folder-relative and `shared:/` paths, serve with a content hash in the URL so screens can cache — `/assets/scenes/<scene id>/...` and `/assets/shared/...` (`infrastructure/assets`)
- [ ] Battle map format v0 — grid settings + ordered layers: image layers and entity layers with instance groups (one asset, many positions); new layer kinds (walls, lights, sound) are added with the feature that needs them

**Client**

- [x] Device / TV view: fullscreen Pixi.js renderer driven by server screen state (`/views/devices/`, reconnects on its own; `?device=` / `?type=` override the device ID / hardware type)
- [ ] Layer system in `src/lib/2d` (image layers, grid, entities, overlays) and the render loop; instance groups rendered with Pixi instancing (shared texture, many positions)
- [x] Views: `battlemap` (background + grid), `scenery` (image or video), `splash`, `blank` — `src/lib/2d/screen-renderer.ts`
- [x] GM view v0: list scenes, activate one, switch variant (`/views/game-master/`)
- [~] Device pairing page (`/devices/pair`): generates and persists the device ID (done, in `$lib/devices/identity.ts`, used by the device view), shows the 4-digit pairing code full screen, renews it on expiry, stores the device token once paired
- [ ] Devices admin page (`/devices`): type in a pairing code + device name, list paired and connected devices (ID, type, name), revoke a device

**Unknowns**

- ~~Battle map format v0?~~ Decided: a battle map is a layered document — grid settings + an ordered list of typed layers. v0 has image layers (background and other images) and entity layers, including instance groups (one asset drawn at many positions, rendered via Pixi instancing for performance). The format grows with features: walls, lights, sound, ... are added as new layer kinds when the feature that uses them lands, not designed up front. Storage format follows the M2 config file format decision
- ~~Where assets live?~~ Decided: each scene is a folder holding its config and its own assets; shared folders (any depth of subfolders, user-organized) hold assets reused across scenes. Configs reference scene assets by path relative to the scene folder (`maps/cave.webp`) and shared assets with a prefix (`shared:/monsters/goblins/goblin.png` — exact syntax to confirm when implemented). The server serves them over HTTP with a content hash in the URL for caching
- ~~Device pairing UX~~ — decided: the device shows a 4-digit code, an admin types it in on the devices page
- ~~Device pairing transport, code lifetime, token format?~~ Decided: a ConnectRPC server stream (`DeviceService.Pair`, `anonymous` route) streams pairing codes and ends with the device token. A code is valid only while the stream is open, for at most 5 minutes, and is renewed automatically; confirm attempts are rate-limited. The device token is a JWT backed by a JTI row (reusing the user token / key infrastructure, revoked via the JTI)
- ~~Device token lifetime and refresh?~~ Decided: pairing issues a device access token (8 h) and a device refresh token (30 days); the device refreshes on its own before expiry. A device unused for more than 30 days has to pair again
- ~~Allowed `X-Device-Type` values?~~ Decided: `screen`, `table`, `tablet`, `phone`, `desktop`, `other`; unknown values are stored as `other`, never rejected. Metadata only, not used for authorization
- ~~Which room is active when there's more than one?~~ Decided: a running instance serves exactly one game table, so there is exactly one room (the table's devices and target bindings); no room selection and no room ID flag

---

## M2 — Scene authoring

**Goal**: a GM can create and edit scenes, variants and the room setup from the app, and import existing maps, without touching JSON files.

- [ ] Config files in YAML or JSON (by extension) with a `version` field and migrations on load; file models translated to `scenes/v1` proto messages for the API / screens
- [ ] Scene editor: entities, base instructions per target, variants
- [ ] Room editor: devices, target bindings
- [ ] Asset upload and management
- [ ] Map import: Universal VTT (`.dd2vtt`) and Dungeondraft into the Mechanus battle map format (`mechanus maps import` is a stub today)
- [ ] Preview a scene / variant without activating it on the room
- [ ] Scene state keyed by campaign + scene, so the same scene config keeps separate state in each campaign
- [ ] Live reload of hand-edited config / asset files: watch the folders, reload changed configs, re-instruct devices when the active scene changes, report invalid files in the GM view and keep the last good version

**Unknowns**

- ~~In-app editor only, or also edit config files directly?~~ Decided: both. The in-app editor is the primary tool; config and asset files stay hand-editable, the server watches the folders and reloads on change (re-instructing devices if the active scene changed). An invalid file is reported in the GM view and the last good version stays in use
- ~~File format for configs?~~ Decided: YAML and JSON are both supported (picked by file extension). Every file carries a `version` field and old versions are migrated on load. The file models are translated to proto messages for network communication; the proto is the wire contract, not the file schema
- ~~Are scenes shareable between campaigns?~~ Decided: scenes are campaign-independent configuration — a camp site or tavern can be used by any campaign. Scene state is saved per campaign (keyed by campaign + scene)

---

## M3 — Live play on the battle map

**Goal**: a full combat encounter can be run on the table: tokens move, fog is revealed, initiative is tracked, and players see only what they should.

- [ ] Entity tokens on the battle map, positions stored in scene state
- [ ] Screen WebSocket messages for pan, zoom, token moves
- [ ] GM controls: reveal / hide / spawn / remove entities
- [ ] Token control: `controlled_by` users per entity (default: the player's own character), GM assigns / shares control, live movement lock for players
- [ ] Fog of war and reveal, stored in scene state
- [ ] Overlays: `initiative` tracker, `entities` portraits
- [ ] Generic game primitives: campaign-defined resource bars on tokens (with visibility), condition tags with icons and round counters, hand-entered initiative order with turn / round tracking, free-form entity stats
- [ ] Info board view (handouts, initiative order, quest log)
- [ ] Player view with limited visibility: derived per user from battle map state (hidden / fogged entities stripped server-side, camera on their character), overridable by a scene's `players` target
- [ ] Reset a scene's state back to its config
- [ ] Touch input on screens: per-device `touch` mode (`none` / `view` / `control` acting as a configured role or user), live GM touch lock

**Unknowns**

- ~~Player views: a target per player, or one derived view?~~ Decided: hybrid. By default every player gets a view derived server-side from the battle map state, filtered per user (hidden / fogged entities are never sent), camera on their own character. Scenes can optionally instruct a `players` target to override what player screens show (handouts, custom views) for special moments
- ~~Who may move tokens?~~ Decided: per-entity controllers. Each entity has a `controlled_by` list of users (a player's character defaults to that player); the GM can always move everything, can assign or share control (summons, companions, mounts), and can toggle a live movement lock that freezes player movement
- ~~How much rules knowledge belongs in Mechanus?~~ Decided: generic primitives, no rules engine and no dice. Campaign-defined resource bars (HP, mana, ...; current / max, visibility everyone / owner / GM), user-defined condition tags with icons and an optional round counter, an initiative order whose values are entered by hand (GM sorts and advances turns / rounds), and free-form key/value stats on entities
- ~~Touch input on a table-mounted TV?~~ Decided: a per-device `touch` setting in the room config — `none`, `view` (pan, zoom, ping, measure; never changes game state; the default) or `control` (moves tokens, acting as a configured role or user). The GM can lock touch input live

---

## M4 — The room comes alive

**Goal**: activating or switching a scene also sets the lights, starts the right music, and triggers custom devices.

- [ ] Driver framework: integration config sets (credentials, bridge / broker addresses), health / connection status per device
- [ ] Lights: Philips Hue
- [ ] Lights: Home Assistant
- [ ] Lights: Matter (native Go controller: commissioning, fabric management)
- [ ] Music: Spotify
- [ ] Music: iTunes / Apple Music
- [ ] Custom devices: MQTT — topic from device config (default `mechanus/devices/<device-id>/set`), retained JSON envelope `{scene, variant, target, payload}`, scene events on `mechanus/events/scene`, availability on `mechanus/status`
- [ ] Third-party event API (JSON WebSocket, scoped API tokens created by an admin): external software subscribes to scene events (scene activated, variant changed, entity changes, instructions for its targets) and sends events back into Mechanus
- [ ] Handle released targets (`Transition.Released`) — reset, stop, or keep
- [ ] GM live controls: adjust lights / volume mid-scene (`SceneService.Adjust`)

**Unknowns**

- ~~Which integration first?~~ Decided: no fixed order, all integrations are equally important; the driver framework comes first since every integration builds on it. A third-party event API (WebSocket / gRPC) joins them, so external software can integrate without a dedicated driver
- Auth flows that expect a public callback URL (Spotify OAuth) on a LAN-only server. Options considered: project-hosted static HTTPS relay page that bounces the code to the LAN server (PKCE, user's own client ID), loopback redirect + paste the URL back, delegate to Home Assistant, or require configured TLS — TBD
- ~~Is Home Assistant enough to cover Hue / Matter?~~ Decided: no — Home Assistant, Philips Hue and Matter are three separate native integrations. Matter means implementing a controller in Go (commissioning, fabric management), with no mature Go library to lean on today
- ~~Custom device payload conventions?~~ Decided: scenes only hold the opaque payload per target; the topic comes from the device config (default `mechanus/devices/<device-id>/set`). Mechanus wraps it in a JSON envelope `{scene, variant, target, payload}` and publishes it retained. Scene events go to `mechanus/events/scene` and availability to `mechanus/status` (last will). The envelope and event schema are shared with the third-party event API
- ~~Third-party event API transport and auth?~~ Decided: a JSON WebSocket only (no ConnectRPC variant), using the event schema shared with MQTT. Third-party software authenticates with admin-created, revocable, scoped API tokens (JWT + JTI)
- Third-party event API scopes: what incoming events may do (candidates: `events:read` subscribe, `events:write` custom events shown in the GM view, `scenes:control` activate scenes / switch variants / adjust) — TBD

---

## M5 — Sessions & campaigns

**Goal**: campaigns, players and play sessions are managed in the app, and a campaign's session continues exactly where it left off — across server restarts and campaign switches.

- [x] Campaign overview and detail pages: users see the campaigns they are in, admins see all; users can create a campaign
- [ ] Campaign management UI: edit, delete campaigns; invite / manage players
- [ ] Characters linked to scene entities (`Entity.CharacterID`)
- [ ] Remember the active scene across server restarts
- [ ] Session object: one continuous, never-ending session per campaign holding the active scene and play state; switching campaign switches session

**Unknowns**

- ~~Is a "session" a first-class object?~~ Decided: yes, but it never stops — a session is the continuous, ongoing game (one per campaign), holding the active scene and play state. There is no start / end lifecycle; switching campaigns switches session and returning resumes where it left off
- ~~Should scene state keep history?~~ Decided: latest only — one saved state per campaign + scene, overwritten on exit; no snapshots or event log

---

## M6 — Polish & release

**Goal**: someone else can install Mechanus on a home network and use it.

- [ ] Client discovers the server via mDNS (server side already announces)
- [ ] Packaging: single binary releases for Linux, Windows and macOS (amd64 + arm64), multi-arch Docker image (amd64 + arm64), setup docs; all platforms tested in CI
- [ ] About page content and licenses (`client/src/routes/about`)
- [ ] macOS configuration paths documented (`docs/configuration/config-file.md`)
- [ ] User-facing documentation: setting up a room, writing scenes

**Unknowns**

- ~~Target platforms?~~ Decided: Linux, Windows and macOS (amd64 + arm64) and a multi-arch Docker image are all fully supported and tested — covers Raspberry Pi, NAS, mini PC and a GM's laptop. Cross-compiling is easy since SQLite is pure Go (no cgo)

---

## Parking lot

Ideas that don't belong to a milestone yet.

- TBD
