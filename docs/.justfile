# OpenSpec recipes. Imported by the root .justfile as `mod spec 'docs/.justfile'`,
# so `just spec <recipe>` runs here in docs/ — the only place the openspec CLI finds
# its root (docs/openspec/). See docs/openspec/README.md.

set windows-shell := ["powershell.exe", "-c"]

export OPENSPEC_TELEMETRY := "0"

default:
  just --list

# List open changes with task progress
[group('openspec')]
list:
  openspec list

# List specs (the capabilities the system already has)
[group('openspec')]
specs:
  openspec list --specs

# Interactive dashboard of specs and changes
[group('openspec')]
view:
  openspec view

# Show a change or spec, e.g. `just spec show add-fog-of-war`
[group('openspec')]
show item *args:
  openspec show {{item}} {{args}}

# Artifact status of a change
[group('openspec')]
status change:
  openspec status --change {{change}}

# Validate every spec and change (strict, like CI should)
[group('openspec')]
validate *args='--all --strict':
  openspec validate {{args}}

# Scaffold a new change, e.g. `just spec new add-dice-tray`
[group('openspec')]
new name:
  openspec new change {{name}}

# Archive a finished change and merge its deltas into the specs
[group('openspec')]
archive name *args:
  openspec archive {{name}} {{args}}

# Regenerate the Claude Code skills and move them to the repo root (.claude/skills, .agents/skills)
[group('openspec')]
[windows]
skills:
  openspec init . --tools claude --scope project --force --no-animation
  New-Item -ItemType Directory -Force ../.claude/skills, ../.agents/skills | Out-Null
  Copy-Item -Recurse -Force .claude/skills/* ../.claude/skills/
  Copy-Item -Recurse -Force .agents/skills/* ../.agents/skills/
  Copy-Item -Force .agents/skills/.openspec-target ../.agents/skills/
  Remove-Item -Recurse -Force .claude, .agents

# Regenerate the Claude Code skills and move them to the repo root (.claude/skills, .agents/skills)
[group('openspec')]
[unix]
skills:
  openspec init . --tools claude --scope project --force --no-animation
  mkdir -p ../.claude/skills ../.agents/skills
  cp -R .claude/skills/. ../.claude/skills/
  cp -R .agents/skills/. ../.agents/skills/
  rm -rf .claude .agents
