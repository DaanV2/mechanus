---
name: pr-creation
description: Use this whenever the user asks to open, create, or ship a pull request for this repo (mechanus) — phrases like "open a PR", "create a pull request", "ship this branch", "put this up for review", or "make a PR for these changes" should all trigger it, even if the user doesn't say "draft". The hard rule this skill enforces is that every PR it creates is opened as a draft (`gh pr create --draft`) unless the user explicitly says to open it ready for review — never assume readiness on the user's behalf.
---

# Creating pull requests for mechanus

Opening a PR notifies reviewers and kicks off CI across multiple workflows (`checks.yaml`, `client-checks.yaml`, `server-checks.yaml`, `proto-checks.yaml`, `docker-checks.yaml`, `integration-tests.yaml`). That's enough of a public, hard-to-undo action that this repo's convention is: **every PR starts as a draft**. A draft still runs CI and lets reviewers see the diff, but it signals "not ready yet" and avoids pulling reviewers in before the branch is actually done. Only skip `--draft` if the user explicitly says the PR should be ready for review immediately.

## Steps

1. **Gather context before writing anything.** Run these together (they're independent):
   - `git status` — confirm there's nothing uncommitted that should be in the PR, and no untracked files that shouldn't be.
   - `git diff` — see staged/unstaged changes.
   - Check whether the branch tracks a remote and is up to date (`git status -sb` or `git rev-parse --abbrev-ref --symbolic-full-name @{u}` wrapped in a check), so you know whether a push is needed.
   - `git log main..HEAD` (or the actual base branch) and `git diff main...HEAD` — the **full** set of commits going into the PR, not just the latest one. A branch with several commits needs the PR body to reflect all of them, not just the tip.

2. **Draft the title and body.**
   - Title: Conventional Commits format (`type(scope): short imperative summary`), under ~70 characters — same as this repo's commit messages (see `git log --oneline -20` for examples like `fix(auth): ...`, `chore(deps): ...`, `feat(scenes): ...`). Common types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`. Put detail in the body, not the title.
   - Body: use a HEREDOC so multi-line text and special characters survive the shell untouched — never try to inline a multi-line `-b` string.
   - Structure:
     ```
     ## Summary
     <1-3 bullet points on what changed and why>

     ## Test plan
     - [ ] <specific, checkable steps a reviewer or CI would run>
     ```
   - If this session's context includes a PR attribution footer (Claude Code's own instructions specify this — check for a line like `🤖 Generated with [Claude Code](https://claude.com/claude-code)`), end the body with it. Don't invent a footer if none was specified.

3. **Push and create, in that order, only after confirming with the user if either is a new action:**
   - If the branch has no upstream yet or has unpushed commits, push it (`git push -u origin <branch>` or `git push`) — but only do this if the user has already asked you to open a PR; opening a PR implies pushing the branch that backs it, so a second confirmation isn't needed for the push itself.
   - Create the PR **as a draft**:
     ```bash
     gh pr create --draft --title "..." --body "$(cat <<'EOF'
     ## Summary
     - ...

     ## Test plan
     - [ ] ...

     🤖 Generated with [Claude Code](https://claude.com/claude-code)
     EOF
     )"
     ```
   - If the user explicitly said the PR should be ready for review (not draft), omit `--draft`. Otherwise always include it — don't ask permission for this part, just do it; the draft rule is the whole point of this skill.

4. **Report back** with the PR URL (`gh pr create` prints it; also available via `gh pr view --json url`).

## What NOT to do without being asked

- Don't mark a draft PR as ready for review (`gh pr ready`) — that's a separate, explicit user request.
- Don't merge, enable auto-merge, or set auto-merge options.
- Don't force-push an existing PR's branch.
- Don't add reviewers, assignees, or labels unless asked — `auto-label-pr.yaml` already labels PRs automatically on open.
- Don't push to `main` or any protected branch directly.

If a `gh pr create` call fails because a PR already exists for the branch, don't create a duplicate — tell the user and offer to update the existing one (`gh pr edit`) instead.
