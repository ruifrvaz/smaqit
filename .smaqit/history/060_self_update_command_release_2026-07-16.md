# Self-Update Command Release

**Date:** 2026-07-16
**Session Focus:** Add a `smaqit update` self-update command to the installer, release it as v1.4.0, and troubleshoot Claude Code VS Code extension slash-command discovery
**Tasks Referenced:** None (no task file — direct feature request)
**Tasks Completed:** N/A

---

## Actions Taken

### Session Start
- Ran `smaqit.session-start`: loaded README, CONTRIBUTING, copilot-instructions, session 059 history, PLANNING.md, and compendium
- Flagged two active tasks (071 Q&A agent/skill, 074 principle taxonomy update) as already resolved in the codebase but not yet moved to Completed in PLANNING.md
- Flagged task 070's priority discrepancy (High in the task file vs Low in the PLANNING.md table)

### `smaqit update` Command — Research and Plan
- Investigated `~/projects/smaqit-extensions/installer/main.go`'s existing self-update mechanism: GitHub releases API fetch, semver comparison, platform-matched asset download, atomic binary replace (same-filesystem fallback), and `.smaqit/`-aware re-init — all pure stdlib, no new dependencies
- Confirmed smaqit's own release pipeline (`.github/workflows/post-merge-release.yml`) produces compatible asset names (`smaqit_<os>_<arch>[.exe]`, 5-platform matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- Entered plan mode; wrote and got approval for a plan to port the mechanism into a new `installer/update.go`, adapted to smaqit's repo (`ruifrvaz/smaqit`) and binary name

### `smaqit update` Command — Implementation and Verification
- Created `installer/update.go`: `runUpdate()`, `fetchLatestRelease()`, `compareVersions()`, `downloadBinary()`, `replaceBinary()`, `copyFile()`, `checkAndReInit()`
- Wired `case "update"` into `installer/main.go`'s command switch, plus `printUsage()` and `cmdHelp()` text
- Added `smaqit update` row to README's CLI commands table
- Verified end-to-end against the real `v1.3.1` GitHub release (which existed on GitHub already, meaning the user had pushed it from their own machine since session 059): built a fake `1.0.0` binary, ran `update` in a scratch dir, confirmed real download + atomic replace (`smaqit version` → `v1.3.1` after), and confirmed the `.smaqit/`-present re-init path triggers `cmdInit`'s existing conflict-confirmation flow correctly

### Commit and Local Release (v1.4.0)
- Discovered uncommitted leftovers from session 059 (`.smaqit/compendium.md`, an untracked history file) and committed them separately first
- Committed the `smaqit update` feature as its own commit
- Ran release-analysis reasoning: MINOR bump (new capability, no breaking change) → `v1.4.0`
- Asked the user how to handle the push, since SSH still doesn't work in this sandbox (same blocker as session 059) but `gh` CLI has an authenticated HTTPS token with at least read access — user chose **local-only, no push**
- Updated `CHANGELOG.md` (`[Unreleased]` → `[1.4.0]`), also fixed a missing `[1.3.1]` comparison link that session 059's release had skipped
- Bumped `installer/main.go`'s `Version` fallback to `1.4.0`
- Committed `Release v1.4.0`, created annotated local tag `v1.4.0` — did not push

### Claude Code Slash Command Troubleshooting
- User asked why slash commands don't appear when typing `/` in the Claude Code VS Code extension chat
- Explained that this repo (`smaqit`'s own source) has no `.claude/commands/` or `.claude/agents/` because those are compiled/gitignored build artifacts (`installer/commands-claude/`, `installer/agents-claude/`) that only get installed into a *target* project via `smaqit init` — this repo only ships its own dev-tooling `.claude/settings.json` + `.claude/hooks/`
- Clarified that even after `smaqit init`, only 4 of 9 agents become Claude Code slash commands (development/deployment/validation/qa) — the five spec agents intentionally have no command file, mirroring Copilot's `user-invocable:false` boundary
- User reported running `smaqit init` on another project but still not seeing commands; researched via WebSearch/WebFetch
- Found official guidance (`Developer: Reload Window` is the documented first troubleshooting step) plus a known unresolved GitHub issue (#9518) reporting the VS Code extension failing to detect `.claude/commands/` specifically on **Linux/WSL** — the user's exact platform — closed as "not planned" with no fix
- Gave a troubleshooting order: reload window → full VS Code restart → verify actual file contents → fall back to running the CLI (`claude`) in an integrated terminal, which several of the linked bug reports suggest works when the extension panel doesn't

---

## Problems Solved

- **Self-update had no existing precedent in this repo**: fully solved by porting the proven, dependency-free mechanism from `smaqit-extensions` rather than designing one from scratch.
- **Uncommitted session 059 artifacts**: found and committed separately before starting new feature work, keeping history clean.
- **Missing v1.3.1 changelog comparison link**: caught and fixed as a side effect of preparing the v1.4.0 release.
- **No SSH access in sandbox (recurring)**: this time offered a concrete alternative (gh CLI's authenticated HTTPS token) rather than just reporting the blocker — user declined it in favor of the established "push from your own machine" pattern.

---

## Decisions Made

- **Self-update lives in its own file (`installer/update.go`)**, not appended to the already-large `main.go` — keeps the diff self-contained and the logic easy to locate.
- **No new dependencies** — the ported mechanism uses only Go stdlib (`net/http`, `encoding/json`, `runtime`, etc.), matching `smaqit-extensions`' approach exactly.
- **Local-only release, no automated push** — user explicitly chose not to use the `gh` HTTPS token workaround for pushing to the real repo, preferring to push from their own machine as in session 059.
- **v1.4.0, not v1.3.2** — new capability (self-update command) is a MINOR bump per semver, not a patch.

---

## Files Modified

### New
- `installer/update.go`
- `.smaqit/history/059_release_v1.3.1_and_instructions_files_2026-07-16.md` (backfilled from session 059, previously uncommitted)

### Modified
- `installer/main.go` — `case "update"` wiring, usage/help text, `Version = "1.4.0"`
- `README.md` — `smaqit update` row in CLI commands table
- `CHANGELOG.md` — `[1.4.0]` section, fixed missing `[1.3.1]` comparison link
- `.smaqit/compendium.md` — backfilled from session 059 (previously uncommitted)

---

## Next Steps

- **Push required**: `git push origin main && git push origin v1.4.0` from a machine with GitHub SSH access — triggers the release workflow (5 platform binaries + public GitHub Release from the `[1.4.0]` CHANGELOG section). Local commits ready: `1eccf76` (docs), `e99fc6c` (feat), `d9cee43` (release prep) — tag `v1.4.0` created locally.
- **Slash command issue unresolved**: user still needs to try `Developer: Reload Window` / VS Code restart on the other project, or fall back to the CLI in an integrated terminal, and report back whether either works.
- **PLANNING.md cleanup still pending** (flagged at session start, not actioned this session): tasks 071 and 074 both appear already complete in the codebase and should be moved to the Completed table; task 070's priority discrepancy (file says High, table says Low) should be reconciled.

---

## Session Metrics

- **Date:** 2026-07-16
- **Commits:** 3 local (`1eccf76`, `e99fc6c`, `d9cee43`) — none pushed
- **New tag:** `v1.4.0` (local only)
- **New installer capability:** `smaqit update` self-update command
- **Files changed:** 6 (2 new, 4 modified across the two work threads)
