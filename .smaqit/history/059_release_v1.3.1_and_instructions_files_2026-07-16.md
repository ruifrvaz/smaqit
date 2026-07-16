# Release v1.3.1 and AGENTS.md/CLAUDE.md Support

**Date:** 2026-07-16
**Session Focus:** Version and release the Claude Code migration (session 058); add AGENTS.md/CLAUDE.md as a new dual-target divergence point
**Tasks Referenced:** None (continuation of session 058, no task file)
**Tasks Completed:** N/A

---

## Actions Taken

### Versioning and Local Release (initially v1.3.0)
- Ran release-analysis reasoning against `git log v1.2.0..HEAD`: MINOR bump (new capability, no breaking removal) → `v1.3.0`
- Updated `CHANGELOG.md` (`## [Unreleased]` → `## [1.3.0]`) and `installer/main.go`'s `Version` fallback
- Committed `82a9e2c "Release v1.3.0"`, created annotated local tag `v1.3.0`
- Built and verified a local binary (`installer/dist/smaqit-dev`) reporting `v1.3.0` via clean `git describe`
- Attempted to push — failed: sandbox has no SSH key access to `git@github.com:ruifrvaz/smaqit.git` (`Permission denied (publickey)`). User must push from a machine with their own GitHub SSH access.

### Installation Instructions Q&A
- Explained current `smaqit init` behavior: one unified command installs both `.github/` (Copilot) and `.claude/` (Claude Code) artifacts together — no `--platform` flag exists to select just one

### AGENTS.md / CLAUDE.md — New Divergence Point
- User identified a third, previously-unaddressed divergence point: project-level "instructions" files (`copilot-instructions.md` vs `CLAUDE.md` vs `AGENTS.md`), and noted the installer doesn't currently write any of them
- Researched via subagents (official docs, not assumptions) and confirmed:
  - VS Code Copilot Chat reads `AGENTS.md` natively (`chat.useAgentsMdFile` setting) alongside `.github/copilot-instructions.md`
  - GitHub Copilot coding agent (cloud) reads `AGENTS.md` natively since 2025-08-28
  - Claude Code does **not** read `AGENTS.md` on its own — official docs: *"Claude Code reads CLAUDE.md, not AGENTS.md"* — recommended pattern is a `@AGENTS.md` import line inside `CLAUDE.md` (cross-platform-safe) or a symlink (breaks on Windows without admin)
- Discovered `.smaqit/templates/copilot-instructions.template.md` is not dead code as previously assumed — it belongs to a separate dev-tooling skill (`.github/skills/smaqit.project-init`, part of a broader "smaqit-extensions" meta-scaffolding system for bootstrapping new repos), unrelated to the shipped `smaqit init` product flow. Left untouched; new templates created instead.
- Designed and built: `templates/AGENTS.md.template` (real content: scaffolding-paths-to-ignore + smaqit usage summary) and `templates/CLAUDE.md.template` (one line: `@AGENTS.md`)
- Added `installInstructionsFile()` to `installer/main.go`: creates the file if absent, appends smaqit's marker-wrapped section if the file exists without that marker, no-ops if already installed (idempotent) — never overwrites existing user content
- Wired into `cmdInit`, `installer/Makefile`'s `prepare` target updated to copy the two new templates
- Verified three scenarios end-to-end: fresh install (both files created), repeat `smaqit init` (idempotent, no duplication), and pre-existing user-authored `AGENTS.md`/`CLAUDE.md` (smaqit's section appended below untouched user content)
- Committed as `e04300f`, landing after the `v1.3.0` tag

### Consolidating to v1.3.1
- User requested bundling everything into a single `v1.3.1` release rather than shipping `v1.3.0` separately
- First attempt to delete+recreate the `v1.3.0` tag was blocked by the Claude Code auto-mode safety classifier (treats tag deletion/recreation as touching potentially-published history without explicit per-operation approval)
- User then explicitly authorized the re-tag; deleted local `v1.3.0` tag, renamed the CHANGELOG section `[1.3.0]` → `[1.3.1]`, bumped `installer/main.go` version to `1.3.1`, committed (`222a0e2 "Release v1.3.1"`), created annotated tag `v1.3.1`
- Rebuilt local binary, confirmed clean `v1.3.1` via `git describe` and `smaqit version`

---

## Problems Solved

- **No SSH access in sandbox**: Push to `origin` fails with publickey rejection. Not something to work around — user pushes from their own machine.
- **Safety classifier blocked tag re-creation**: Correctly refused an ambiguous destructive git operation (deleting a tag that might already be public) without explicit per-operation user approval. Resolved by asking the user directly rather than attempting a workaround; user then explicitly authorized it.
- **Misidentified dead template file**: `.smaqit/templates/copilot-instructions.template.md` looked unused (no installer reference) but actually belongs to a separate `smaqit.project-init` dev-tooling skill. Avoided incorrectly repurposing or deleting it by tracing its actual reference before acting.

---

## Decisions Made

- **AGENTS.md is the single content file; CLAUDE.md is a thin `@AGENTS.md` import** — matches confirmed platform support exactly (both Copilot surfaces read AGENTS.md; Claude Code requires the explicit pointer). No third `.github/copilot-instructions.md` file needed since Copilot already reads AGENTS.md directly.
- **Create-if-absent, append-if-exists, idempotent via HTML-comment marker** (`<!-- smaqit:instructions:begin/end -->`) — never overwrites a user's own instructions file, confirmed by user preference over both "skip if exists" and "always overwrite" alternatives.
- **`.smaqit/templates/copilot-instructions.template.md` is out of scope** — belongs to the separate `smaqit-extensions`/`smaqit.project-init` dev-tooling system, not the shipped product installer. Not touched.
- **v1.3.0 was fully retired before ever being pushed** — only `v1.3.1` exists in tag history, bundling the Claude Code migration (session 058) and the AGENTS.md/CLAUDE.md work together as one release.

---

## Files Modified

### New
- `templates/AGENTS.md.template`, `templates/CLAUDE.md.template`

### Modified
- `CHANGELOG.md` — `[1.3.1]` section covering both the dual-target migration and AGENTS.md/CLAUDE.md support
- `installer/main.go` — `Version = "1.3.1"`, `installInstructionsFile()` helper, embeds for the two new templates, `cmdInit` wiring
- `installer/Makefile` — `prepare` target copies the two new templates
- `README.md` — `smaqit init` description updated to mention `.claude/`, `AGENTS.md`, `CLAUDE.md`

---

## Next Steps

- **Push required**: `git push origin main && git push origin v1.3.1` from a machine with GitHub SSH access — this triggers `.github/workflows/post-merge-release.yml`, building 5 platform binaries and publishing the public GitHub Release from the `[1.3.1]` CHANGELOG section
- No selective single-platform install flag exists yet (`smaqit init` always installs both `.github/` and `.claude/` artifacts together) — flagged to user as a possible future addition, not yet requested

---

## Session Metrics

- **Date:** 2026-07-16
- **Commits:** 3 (`82a9e2c`, `e04300f`, `222a0e2` — net result: single `v1.3.1` tag, `v1.3.0` never published)
- **New installer capability:** `AGENTS.md`/`CLAUDE.md` installation (create/append/idempotent)
- **Local binary built and verified:** `installer/dist/smaqit-dev` → `smaqit v1.3.1`
- **Blocked (external, not resolved this session):** push to remote (no SSH access in sandbox)
