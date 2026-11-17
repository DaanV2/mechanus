# Roadmap

What still has to be built to get Mechanus from "foundation" to "the whole game room runs from the browser". Milestones are ordered by dependency, not by date.

The detail lives in [OpenSpec](./openspec/README.md):

- **What the system does today** → [`openspec/specs/`](./openspec/specs/) (one spec per capability)
- **What is still to build** → [`openspec/changes/`](./openspec/changes/) (one change per feature: proposal, delta specs, design with open questions, tasks)

This file is only the milestone map: it says which change belongs where and in what order. When a change is archived, remove it here; when a milestone has no open changes left, mark it done.

```ascii
 M0 Foundation ──► M1 First scene ──► M3 Live play ──► M5 Sessions & campaigns
                        │                 ▲
                        ├──► M2 Authoring ┘
                        │
                        └──► M4 The room comes alive

 M6 Polish & release runs alongside everything
```

Open decisions (formerly `TBD`) are listed under **Open Questions** in each change's `design.md`.

## M0 — Foundation

**Goal**: the building blocks every later milestone relies on. Done, except:

- [`move-security-checks-into-handlers`](./openspec/changes/move-security-checks-into-handlers/) — `DeviceService` off `grpc.PolicyInterceptor`, entry / private handler split

Specs: [user-auth](./openspec/specs/user-auth/spec.md), [user-management](./openspec/specs/user-management/spec.md), [screen-connection](./openspec/specs/screen-connection/spec.md), [distribution](./openspec/specs/distribution/spec.md).

Not yet a change: per-screen access ownership (users currently open every screen) waits on a screen / room ownership model and a GM role.

## M1 — First scene on the table

**Goal**: the GM picks a scene and a TV registered as a device shows its battle map or scenery; switching variant updates the TV. **Done.**

Specs: [scene-control](./openspec/specs/scene-control/spec.md), [screen-views](./openspec/specs/screen-views/spec.md), [battle-maps](./openspec/specs/battle-maps/spec.md), [device-pairing](./openspec/specs/device-pairing/spec.md).

## M2 — Scene authoring

**Goal**: a GM can create and edit scenes, variants and the room setup from the app, and import existing maps, without touching JSON files.

- [`add-versioned-config-files`](./openspec/changes/add-versioned-config-files/) — YAML / JSON, `version` + migrations, live reload
- [`add-scene-editor`](./openspec/changes/add-scene-editor/) — scene and room editors, preview
- [`add-asset-management`](./openspec/changes/add-asset-management/) — upload and manage assets
- [`add-map-import`](./openspec/changes/add-map-import/) — Universal VTT and Dungeondraft
- Scene state keyed by campaign + scene → part of [`add-campaign-sessions`](./openspec/changes/add-campaign-sessions/) (M5)

## M3 — Live play on the battle map

**Goal**: a full combat encounter can be run on the table: tokens move, fog is revealed, initiative is tracked, and players see only what they should.

- [`add-token-control`](./openspec/changes/add-token-control/) — `controlled_by`, GM control assignment, movement lock
- [`add-fog-of-war`](./openspec/changes/add-fog-of-war/)
- [`add-game-primitives`](./openspec/changes/add-game-primitives/) — resource bars, conditions, initiative, stats; initiative / portrait overlays
- [`add-info-board-view`](./openspec/changes/add-info-board-view/)
- [`add-player-view`](./openspec/changes/add-player-view/)
- [`add-screen-touch-input`](./openspec/changes/add-screen-touch-input/)

## M4 — The room comes alive

**Goal**: activating or switching a scene also sets the lights, starts the right music, and triggers custom devices. Driver framework, device statuses and MQTT are done ([device-drivers](./openspec/specs/device-drivers/spec.md)). No fixed order between integrations.

- [`add-hue-light-driver`](./openspec/changes/add-hue-light-driver/)
- [`add-home-assistant-light-driver`](./openspec/changes/add-home-assistant-light-driver/)
- [`add-matter-light-driver`](./openspec/changes/add-matter-light-driver/)
- [`add-spotify-music-driver`](./openspec/changes/add-spotify-music-driver/) — open: OAuth without a public callback URL
- [`add-apple-music-driver`](./openspec/changes/add-apple-music-driver/)
- [`add-event-api`](./openspec/changes/add-event-api/) — open: token scopes
- [`handle-released-targets`](./openspec/changes/handle-released-targets/) — open: reset, stop or keep
- [`add-live-scene-adjust`](./openspec/changes/add-live-scene-adjust/)

## M5 — Sessions & campaigns

**Goal**: campaigns, players and sessions are managed in the app, and a campaign's session continues exactly where it left off. Campaign management is done ([campaigns](./openspec/specs/campaigns/spec.md)).

- [`link-characters-to-entities`](./openspec/changes/link-characters-to-entities/)
- [`add-campaign-sessions`](./openspec/changes/add-campaign-sessions/) — session per campaign, active scene survives restarts

## M6 — Polish & release

**Goal**: someone else can install Mechanus on a home network and use it. Packaging, about page and macOS paths are done ([distribution](./openspec/specs/distribution/spec.md)).

- [`add-client-server-discovery`](./openspec/changes/add-client-server-discovery/)
- [`serve-web-client-by-default`](./openspec/changes/serve-web-client-by-default/) — open: embed vs `web/` folder
- [`add-user-documentation`](./openspec/changes/add-user-documentation/)
- [`polish-client-ui`](./openspec/changes/polish-client-ui/) — styling, contrast, dark mode, page fixes (the former `TODO.md`)

## Parking lot

Changes that don't belong to a milestone yet.

- [`order-screen-state-delivery`](./openspec/changes/order-screen-state-delivery/) — screen messages can arrive out of order (seen in the code, not yet in practice)
