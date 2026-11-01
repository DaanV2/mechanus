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
- [~] User accounts, login, JWT / JTI auth with refresh / logout, roles — see [Auth & tokens](#auth--tokens) for what is still missing (invites, screen access check)
- [x] Admin user list page (`/users`, `UserService.List`)
- [x] ConnectRPC services (users, login) and the screen WebSocket (`Ping`, `InitialSetupRequest`)
- [x] Database backends (SQLite, Postgres, MySQL, in-memory) with migrations
- [x] mDNS announcement of the server on the LAN
- [x] SvelteKit client shell with login / signup / profile and placeholder GM / player / device views
- [~] Scene model: config vs. state, variants, targets, rooms, transitions, `SceneService`, `scenestore` (code + tests, not wired)

### Auth & tokens

What exists: login (`LoginService.Login`) starts a session (one `JTI` row per login) and issues an RS512 access token (`scope: "*"`, 1 h) plus a refresh token (`scope: "refresh"`, 7 days); `Refresh` rotates both, `Logout` revokes the session. The ConnectRPC `Authenticator` interceptor validates tokens and puts the claims on the context; handlers answer `Unauthenticated` vs `PermissionDenied` through `grpc.requireRole`. Admins can list users, create accounts, reset passwords (revoking sessions) and set roles; users can change their own password. Lifetimes, signing key rotation and open sign-up are configurable (`auth.*`), see [authentication.md](./architecture/authentication.md). Device auth is tracked under [M1](#m1--first-scene-on-the-table).

**Bugs**

- [x] `grpc.Authenticator.validateAndInject` panicked (nil pointer) on any malformed, expired or badly signed bearer token — now treated as an anonymous request
- [x] Screen WebSocket never authenticated browser users (browsers can't set `Authorization` on `new WebSocket(...)`) — the client now passes its token as a `mechanus.bearer.<jwt>` subprotocol and the server also reads the `access-token` cookie (`authentication.BearerFromHeaders`)
- [x] Cookie fallback in `grpc.getJwtValue` never matched the URL-encoded `Bearer%20<jwt>` cookie — replaced by `authentication.BearerFromHeaders`, which unescapes it
- [x] `Refresh` accepted any `scope: "*"` access token, so a leaked token could be renewed forever — it now only accepts `scope: "refresh"` tokens, which login issues next to the access token
- [x] Admin password reset (`ResetPassword`) didn't revoke the user's JTI — it now revokes every session of the user
- [x] Tokens leaked into logs (`JWTService.validate` debug log, client `Cookie.get` / `Cookie.set` and the profile page `console.log`) — removed
- [x] Invalid / expired tokens made handlers answer `PermissionDenied` — they now answer `Unauthenticated` (no valid token) vs `PermissionDenied` (missing role)
- [x] `KeyManager.GetSigningKey` picked a random key once more than one existed — it now signs with the newest key (creation time is stored with the key)
- [x] The documented 5-minute expiry leeway never applied — it is now passed to the parser (`JWT_LEEWAY`), which also pins the algorithm to RS512 and checks the audience
- [x] `dbStorage.Delete` passed the key as a raw SQL condition, so deleting a stored item always failed
- [x] The signup page created every account with the password `password` instead of the one typed in

**Missing**

- [~] Account creation: admin-created accounts and the open sign-up flag (`auth.signup-enabled`, default off, on in the dev / test compose files) gating `UserService.Create` are done; still missing: invites (GM generates a link / QR code with expiry, player picks their own credentials) and an admin UI to create accounts
- [x] Logout: `LoginService.Logout` revokes the session (from a given access / refresh token or the request's access token); the nav bar's sign out calls it and clears the cookies
- [x] Proper refresh flow: login returns an access token + a `scope: "refresh"` token; the client refreshes a minute before expiry and once on `Unauthenticated` (`client/src/lib/authentication/session.ts`)
- [x] Client checks `exp` — an expired access token no longer counts as logged in (`userHandler.restore()` refreshes first), and both cookies get a `Max-Age` matching their token
- [x] WebSocket auth for users: `Authorization` header, `mechanus.bearer.<jwt>` subprotocol, or `access-token` cookie
- [ ] The `// TODO: check that screenHandler is allowed to be access by this requester` screen access check — needs the screen / room ownership model from M1 first
- [x] Long-lived WebSocket connections: user connections are closed (1008) when the access token expires, and when a periodic check (`server.websocket.session-check-interval`) finds the session revoked. The client doesn't reconnect with a refreshed token yet
- [x] Per-session JTIs (one per login, so one device can be signed out); revoked JTIs and JTIs unused for longer than the max token lifetime are deleted on startup
- [x] Self-service password change (`UserService.ChangePassword`, revokes other sessions, form on the profile page); admin role management API (`UserService.SetRoles`, role changes reach the token on refresh). No admin UI for roles yet
- [x] Login hardening: per-username lockout after 5 failures for 15 minutes (in memory, per server process), and an unknown username is compared against a dummy hash so it takes as long as a wrong password
- [x] Signing key rotation: new key every `auth.key-rotation` (30 days), old keys are deleted once every token they signed has expired
- [x] Configurable token lifetimes: `auth.access-token-duration`, `auth.refresh-token-duration`
- [x] Cookie hardening: `Max-Age` matching the token expiry, `Secure` when the page is served over HTTPS. The tokens stay in JS-readable `SameSite=Lax` cookies (see Unknowns)
- [~] Optional TLS: `server.tls.cert` / `server.tls.key` (`SERVER_TLS_CERT` / `SERVER_TLS_KEY`) serve HTTPS; still missing: honour `X-Forwarded-Proto` behind a reverse proxy (nothing server-side depends on the scheme yet)

**Tests**

- [x] `grpc.Authenticator`: no header, valid, malformed, expired, tampered, revoked and refresh-scope tokens, and the URL-encoded cookie fallback
- [x] `LoginServiceHandler`: login success, wrong password, unknown user, empty fields, lockout; refresh with valid / revoked / access / garbage tokens; logout by token and by request
- [x] `UserServiceHandler` authorization: `Create` (sign-up on / off), `Get` (self / admin / other / anonymous / invalid token), `List` and `ResetPassword` (admin vs. non-admin), reset revokes sessions, `ChangePassword`, `SetRoles`
- [x] `JWTService` negative cases: expired, within leeway, not-yet-valid, missing `exp`, wrong issuer / audience, unknown / missing / non-string `kid`, tampered signature, JTI belonging to another user, unknown JTI; sessions, refresh and cleanup
- [x] `KeyData` PEM marshal / unmarshal round trip; `KeyManager` loads keys on `AfterInitialize`, reuses the signing key, picks the newest, errors on an unknown `kid`, rotates and retires keys
- [~] WebSocket `authenticate`: user token accepted and bad token rejected (header, subprotocol and cookie), refresh-scope token rejected are covered; still missing: route policies return 401 / 403, device ID / type read from header and query fallback, connections closed on expiry / revocation (needs a way to register a screen in tests)
- [ ] Playwright: logout, session survives token expiry via refresh, revoked session is logged out, WebSocket connects as the logged-in user, change password

**Unknowns**

- ~~Is open sign-up intended?~~ Decided: accounts come from GM invites (link / QR) or admin creation; open sign-up only behind a config flag, off by default
- ~~Where tokens live on the client?~~ Decided: keep the JS-readable cookie; fix the encoding, logging and `Max-Age` bugs rather than moving to `HttpOnly` / in-memory storage. Same-origin cookies are still sent on the WebSocket upgrade, so the server can read the cookie for WebSocket auth
- ~~HTTPS on a LAN server?~~ Decided: plain HTTP by default; optional TLS via config (cert / key paths) and reverse-proxy support (`X-Forwarded-Proto`). Cookies get `Secure` only when the request arrived over HTTPS. Screens need a wake-lock fallback since the Screen Wake Lock API requires a secure context

---

## M1 — First scene on the table

**Goal**: the GM picks a scene in the GM view, and a TV registered as a device shows that scene's battle map (or scenery). Switching variant updates the TV. Everything else can be rough.

**Server**

- [ ] Wire `SceneService` + `scenestore` into `components` (the instance's single room, `Exit` on `BeforeShutdown`)
- [x] Screen state proto: `ScreenState` (view, overlays, layers, entities, grid, camera) is sent on initial setup and via `ScreenHandler.SetState`; entity / camera delta messages defined
- [ ] Screen driver: turn a `ScreenInstruction` + scene entities into screen state and push it over the WebSocket
- [ ] Route screen WebSocket connections by device ID (one screen per screen device)
- [ ] Per-route auth policies (`anonymous`, `device`, `user`, `role:<role>`): authentication is optional at connect, enforced by the route; invalid credentials are rejected instead of treated as anonymous — see [authentication.md](./architecture/authentication.md#device-authentication)
- [ ] Device identity: read `X-Device-ID` / `X-Device-Type` (query fallback `deviceId` / `deviceType` for browsers) and store them in the connection's session info (`screens.ConnectionInfo`, logger / trace attributes)
- [ ] Device registration: a device identifies itself, gets a stable ID, and can be bound to a target
- [ ] Device pairing: `DeviceService.Pair` server stream pushes a 4-digit single-use code (valid while the stream is open, max 5 min, auto-renewed), confirmed by an admin, rate-limited; issues a revocable device JWT backed by a JTI (`Authorization: Device <token>`, 8 h) plus a device refresh token (30 days) the device uses to renew itself — currently any non-JWT connection is accepted
- [ ] Scene control API (`scenes/v1` proto): list scenes, activate, switch variant, get active state
- [ ] Asset serving for scene assets (maps, images, video): resolve scene-folder-relative and `shared:/` paths, serve with a content hash in the URL so screens can cache
- [ ] Battle map format v0 — grid settings + ordered layers: image layers and entity layers with instance groups (one asset, many positions); new layer kinds (walls, lights, sound) are added with the feature that needs them

**Client**

- [ ] Device / TV view: fullscreen Pixi.js renderer driven by server screen state
- [ ] Layer system in `src/lib/2d` (image layers, grid, entities, overlays) and the render loop; instance groups rendered with Pixi instancing (shared texture, many positions)
- [ ] Views: `battlemap` (background + grid), `scenery`, `splash`, `blank`
- [ ] GM view v0: list scenes, activate one, switch variant
- [ ] Device pairing page (`/devices/pair`): generates and persists the device ID, shows the 4-digit pairing code full screen, renews it on expiry, stores the device token once paired
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
