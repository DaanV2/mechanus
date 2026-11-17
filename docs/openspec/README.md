# OpenSpec

Mechanus plans features with [OpenSpec](https://github.com/Fission-AI/OpenSpec) (spec-driven workflow). The milestone order is in [../roadmap.md](../roadmap.md).

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

Lifecycle: propose (or pick an existing change) → apply (work through `tasks.md`) → archive (deltas merge into `specs/`, and the change's line comes out of the roadmap).

CI: [`openspec-checks.yaml`](../../.github/workflows/openspec-checks.yaml) runs `openspec validate --all --strict` on PRs that touch `docs/openspec/`, with the CLI version pinned there.

Requires the OpenSpec CLI (`npm install -g @fission-ai/openspec`, or the standalone installer from openspec.dev). Telemetry is turned off in the just recipes (`OPENSPEC_TELEMETRY=0`).
