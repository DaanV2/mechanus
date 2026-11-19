# OpenSpec

Mechanus plans features with [OpenSpec](https://github.com/Fission-AI/OpenSpec) (spec-driven workflow). This file also holds the [milestones](#milestones): the order changes are built in.

- [`specs/`](./specs/) — what the system does **today**, one folder per capability. Requirements use SHALL / MUST, each with `#### Scenario:` WHEN / THEN cases.
- [`changes/`](./changes/) — work **still to build**. Each change has a `proposal.md` (why / what / milestone), delta specs under `specs/` (`## ADDED Requirements` etc.), an optional `design.md` (decisions and open questions) and a `tasks.md` checklist.
- [`changes/archive/`](./changes/archive/) — finished changes. Archiving merges a change's deltas into `specs/`.
- [`config.yaml`](./config.yaml) — project context and rules that guide the workflows.

## Location

The CLI always looks for a folder named `openspec/` by walking up from the current directory, so the root here is `docs/`. Run the CLI from `docs/`, or use the just module from the repo root:

```bash
just spec list        # open changes and task progress
just spec specs       # capabilities
just spec show add-fog-of-war
just spec validate    # --all --strict
just spec new add-dice-tray
just spec archive add-fog-of-war
just spec skills      # regenerate the Claude Code skills into the repo-root .claude/skills
```

## With Claude Code

The OpenSpec skills are committed in `.claude/skills/openspec-*` (and `.agents/skills/`): `/openspec-propose`, `/openspec-explore`, `/openspec-apply-change`, `/openspec-update-change`, `/openspec-sync-specs`, `/openspec-archive-change`. The skills run `openspec` commands, which must run from `docs/`.

Lifecycle: propose (or pick an existing change) → apply (work through `tasks.md`) → archive (deltas merge into `specs/`, and the change comes off the milestone list below).

CI: [`openspec-checks.yaml`](../../.github/workflows/openspec-checks.yaml) runs `openspec validate --all --strict` on PRs that touch `docs/openspec/`, with the CLI version pinned there.

Requires the OpenSpec CLI (`npm install -g @fission-ai/openspec`, or the standalone installer from openspec.dev). Telemetry is turned off in the just recipes (`OPENSPEC_TELEMETRY=0`).

## Milestones

What still has to be built to get Mechanus from "foundation" to "the whole game room runs from the browser". Milestones are ordered by dependency, not by date. Each proposal names its milestone, and so does the `goal` in its `.openspec.yaml`. When you add or archive a change, update this list; mark a milestone done once it has no open changes.

```ascii
 M0 Foundation ──► M1 First scene ──► M3 Live play ──► M5 Sessions & campaigns
                        │                 ▲
                        ├──► M2 Authoring ┘
                        │
                        └──► M4 The room comes alive

 M6 Polish & release runs alongside everything
```

### M0 — Foundation

**Goal**: the building blocks every later milestone relies on: server, auth, transport, the scene model. Done, except:

- [`move-security-checks-into-handlers`](./changes/move-security-checks-into-handlers/): `DeviceService` off `grpc.PolicyInterceptor`, entry / private handler split
- [`add-screen-wake-lock`](./changes/add-screen-wake-lock/): keep screens awake, with a fallback over plain HTTP

### M1 — First scene on the table

**Goal**: the GM picks a scene and a TV registered as a device shows its battle map or scenery; switching variant updates the TV. **Done.**

### M2 — Scene authoring

**Goal**: a GM can create and edit scenes, variants and the room setup from the app, and import existing maps, without touching JSON files.

- [`add-scene-editor`](./changes/add-scene-editor/): scene and room editors, preview
- [`add-asset-management`](./changes/add-asset-management/)
- [`add-map-import`](./changes/add-map-import/): Universal VTT and Dungeondraft
- Scene state keyed by campaign + scene: part of [`add-campaign-sessions`](./changes/add-campaign-sessions/) (M5)

### M3 — Live play on the battle map

**Goal**: a full combat encounter can be run on the table: tokens move, fog is revealed, initiative is tracked, and players see only what they should.

- [`add-token-control`](./changes/add-token-control/)
- [`add-fog-of-war`](./changes/add-fog-of-war/)
- [`add-game-primitives`](./changes/add-game-primitives/): resource bars, conditions, initiative, stats, initiative / portrait overlays
- [`add-info-board-view`](./changes/add-info-board-view/)
- [`add-player-view`](./changes/add-player-view/)
- [`add-screen-touch-input`](./changes/add-screen-touch-input/)

### M4 — The room comes alive

**Goal**: activating or switching a scene also sets the lights, starts the right music, and triggers custom devices. The driver framework, device statuses and MQTT are done. No fixed order between integrations.

- [`add-hue-light-driver`](./changes/add-hue-light-driver/)
- [`add-home-assistant-light-driver`](./changes/add-home-assistant-light-driver/)
- [`add-matter-light-driver`](./changes/add-matter-light-driver/)
- [`add-spotify-music-driver`](./changes/add-spotify-music-driver/): open question on OAuth without a public callback URL
- [`add-apple-music-driver`](./changes/add-apple-music-driver/)
- [`add-event-api`](./changes/add-event-api/): open question on token scopes
- [`handle-released-targets`](./changes/handle-released-targets/): open question on reset, stop or keep
- [`add-live-scene-adjust`](./changes/add-live-scene-adjust/)

### M5 — Sessions & campaigns

**Goal**: campaigns, players and sessions are managed in the app, and a campaign's session continues exactly where it left off. Campaign management is done.

- [`link-characters-to-entities`](./changes/link-characters-to-entities/)
- [`add-campaign-sessions`](./changes/add-campaign-sessions/): session per campaign, active scene survives restarts

### M6 — Polish & release

**Goal**: someone else can install Mechanus on a home network and use it. Packaging, the about page, macOS paths and the client UI polish are done.

- [`add-client-server-discovery`](./changes/add-client-server-discovery/)
- [`serve-web-client-by-default`](./changes/serve-web-client-by-default/): open question on embedding vs a `web/` folder
- [`add-user-documentation`](./changes/add-user-documentation/)

### Cross-cutting

- [`add-gm-role`](./changes/add-gm-role/): the GM role that several "until a GM role exists" decisions wait on (scene control, invites into a campaign, campaign management, per-screen access)

### Parking lot

Changes that don't belong to a milestone yet.

- [`order-screen-state-delivery`](./changes/order-screen-state-delivery/): screen messages can arrive out of order (seen in the code, not yet in practice)
