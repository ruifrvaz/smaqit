# Self Update Reinit Bug Fix

**Date:** 2026-07-21
**Session Focus:** Diagnose and fix a self-update bug where `smaqit update` silently skipped new skills/scripts after replacing the binary on disk; release v1.5.1 with the fix; plan Task 087 (dynamic stack detection + deploy-skill synthesis).
**Tasks Referenced:** 084, 087
**Tasks Completed:** 084 (documentation closure, findings finalized)

---

## Actions Taken

### Session Start
- Ran `smaqit.session-start`: loaded README, CONTRIBUTING, copilot-instructions, session 060 history, PLANNING.md, and compendium
- Presented project state with 8 active tasks, identified sequencing dependencies between 084/085/086/087 all touching the same file section

### Task 087 Planning
- Invoked `smaqit.task-plan 087` (Mode B — complex task)
- Assessed complexity: Complex (triggers: unresolved dependency chain — 087 rewrites the exact step Tasks 084/086 had just edited)
- Discovered via codebase exploration that 084/086 were already merged into the working tree despite being "In Progress" in PLANNING.md — no sequencing blocker in practice
- Also discovered `smaqit.create-skill` and `smaqit.L2` exist in downstream projects (via `smaqit-adk`) but not in this canonical repo — confirmed the manual-authoring fallback is the realistic default
- Produced full execution plan: replace Phase 4 Step 6's hardcoded two-way stack list with generic stack-spec-driven matching + synthesis procedure, using `description:` frontmatter on installed deploy skills as the matching signal
- Plan approved by user; deferred implementation

### Workspace File
- Created `smaqit.code-workspace` adding a downstream project as a second folder root for multi-repo context

### Stale Embedded Directories Diagnosis
- User reported missing skills/scripts (python-nextjs skill, ownership-guard, plan-guard) in installed projects
- Diagnosed: `installer/skills-copilot/` and `installer/skills-claude/` are gitignored build artifacts regenerated only by `make prepare` — they hadn't been regenerated since Tasks 084–086 landed (Jul 20–21)
- Confirmed via `diff -rq skills/ installer/skills-copilot/`: 5 missing items (python-nextjs skill, `write-vhost.sh`, `ownership-guard.sh`, `sync-secrets.sh`, 4 workflow template assets)
- Root cause: stale-build, not a script bug — `scripts/generate-agents.py`'s `shutil.copytree` was correct; it just needed re-running
- Regenerated with `make -C installer prepare`, rebuilt dev binary (v1.3.1 → v1.5.0-dirty, 54 skill files now), verified via re-diff — only expected `[SMAQIT_SKILLS_DIR]` token substitutions remained differing
- Also confirmed the GitHub Actions release pipeline runs `make prepare` before building, so the published v1.5.0 release binaries were never stale — only the local dev binary was

### `smaqit update` Self-Reinit Bug — Full Diagnosis and Reproduction
- User reported `smaqit update` still didn't install new skills even after typing "yes" to the conflict prompt
- First reproduced the all-or-nothing conflict-gate issue: `detectConflicts()` flags existing files, `cmdInit` gates the entire copy behind a single `Scanln`—new files (no conflicts) also get blocked if the user declines/enters empty response
- User then said they typed "yes", `.claude/` was created, but nextjs skill/guards were still missing — this symptom pointed to a deeper bug
- Confirmed the real installed binary (v1.5.0 on PATH) and the actual GitHub release (v1.5.0 on 2026-07-21, all 5 platform binaries on GitHub) both contain the correct content — fresh `smaqit init` with the real binary installs everything correctly
- **Root cause discovered:** In `installer/update.go`, `runUpdate()` downloads the new binary, replaces it on disk (`replaceBinary`), then immediately calls `checkAndReInit(".")` in the **same still-running process**. Go's `//go:embed` bakes file content into the binary at compile time — replacing the file on disk does nothing to the already-loaded process image in memory. The reinit step silently runs with the **old** binary's stale embedded content. This explains both symptoms exactly:
  - `.claude/` scaffolding "worked" because it existed in the old binary's embed from an earlier release
  - But brand-new v1.5.0 content (python-nextjs skill, `ownership-guard.sh`, `write-vhost.sh`, `sync-secrets.sh`) didn't exist in the old process's embed at all — never written, no error raised
- Built a reproducible test: old-labeled binary (v1.0.0) with deliberately stale skills tree, ran `update` against the real v1.5.0 GitHub release, confirmed the subprocess reinit fixes it

### Self-Reinit Bug Fix
- Added `"os/exec"` import
- Created `reinitWithBinary(binaryPath, dir string)` — re-execs the freshly-downloaded binary on disk as a subprocess (`exec.Command(binaryPath, "init", dir)` with stdin/stdout/stderr wired through)
- Changed the `cmp < 0` (actual update) path to call `reinitWithBinary(currentBin, ".")` instead of `checkAndReInit(".")`
- Preserved `checkAndReInit` for the no-replacement paths (`cmp == 0` / `cmp > 0`) where the in-process embed already matches what's on disk
- Verified end-to-end: built a `1.0.0`-labeled binary with stale skills, ran `update` → subprocess reported `v1.5.0` (not `1.0.0`), confirmed all previously-missing files landed on disk

### Release v1.5.1
- Committed two logical groups: task 084 documentation closure (`28b55c6`) and the update.go fix (`35fd1d1`)
- Release analysis: 1 fix + 1 chore, PATCH severity, suggested v1.5.1
- Updated CHANGELOG.md (`[Unreleased]` → `[1.5.1]` with Fixed and Chore entries, comparison links updated)
- Bumped `installer/main.go` Version to `1.5.1`
- Committed `Release v1.5.1` (`8edec09`), created annotated tag `v1.5.1` (`1590e2e`)
- Pushed both commit and tag to origin — confirmed on remote via `git ls-remote`

### Compact Bug Explanation for smaqit-extensions
- User requested a copy-pasteable version of the bug explanation for fixing the same issue in `smaqit-extensions` (the installer was originally ported from there)

---

## Problems Solved

- **`smaqit update` silently skipped new content after self-replace:** fully diagnosed and fixed — go:embed is compile-time, replacing the binary on disk doesn't change the running process's memory. Fix: re-exec the new binary as a subprocess for reinit.
- **Stale installer embedded directories:** diagnosed, regenerated, documented the `make prepare` requirement — not a code bug, a build-process gap.
- **All-or-nothing conflict prompt in `cmdInit`:** reproduced and documented — new files (no conflict) are blocked when user declines the prompt for existing-file overwrites. Left unfixed in this session (separate from the self-update bug).
- **Cross-session context:** confirmed the GitHub Actions release pipeline already runs `make prepare`, so published binaries were never stale — only local dev builds were affected.

---

## Decisions Made

- **reinitWithBinary subprocess, not a separate flag or flag file:** simpler, self-contained, no state on disk to manage, and the subprocess version string in output provides immediate confirmation the correct binary was used.
- **`checkAndReInit` preserved for no-replacement paths:** the `cmp == 0` / `cmp > 0` branches have no new binary on disk to re-exec, and their in-process embed matches what's on disk — re-running in-process is correct and cheaper here.
- **Task 087 plan deferred:** the plan was approved but implementation intentionally not started — task 087 rewrites the same Phase 4 Step 6 that 084/086 touched, and this session's priority was the self-update bug and release.
- **v1.5.1, not v1.5.0-hotfix:** a conventional PATCH release — single fix, no new features.
- **Workspace file not committed:** `smaqit.code-workspace` is machine-specific, not part of the smaqit product — left untracked.

---

## Files Modified

### New
- `smaqit.code-workspace` — VS Code multi-root workspace (untracked, personal)
- `.smaqit/history/062_self_update_reinit_bug_fix_2026-07-22.md` — this file

### Modified
- `installer/update.go` — new `"os/exec"` import, new `reinitWithBinary()` function, `runUpdate()`'s update path calls it instead of `checkAndReInit()`
- `CHANGELOG.md` — `[1.5.1]` section with Fixed and Chore entries, updated comparison links
- `installer/main.go` — Version bumped to `1.5.1`
- `.smaqit/tasks/084_deploy_target_resolution_provisioning_branch.md` — Findings section finalized, moved from "interim" to final state, documented accepted gap

---

## Next Steps

- **Task 087 implementation:** plan is approved and ready to execute — rewrite Phase 4 Step 6 in `smaqit.new-greenfield-project/SKILL.md` with generic stack-spec-driven matching + synthesis procedure
- **`smaqit-extensions` update bug:** the same self-update reinit bug exists there — apply the same `reinitWithBinary` fix
- **Conflict-prompt redesign:** the all-or-nothing `detectConflicts()` gate in `cmdInit` should separate new-file copies (always proceed) from genuine overwrites (gated) — tracked but not implemented this session
- **PLANNING.md cleanup:** tasks 071 and 074 flagged as possibly already complete in the codebase but still listed as Active — verify and move to Completed
- **Task 070 priority discrepancy:** file says High, PLANNING.md table says Low — should be reconciled

---

## Session Metrics

- **Date:** 2026-07-21
- **Commits:** 3 local (`28b55c6`, `35fd1d1`, `8edec09`) — all pushed
- **New tag:** `v1.5.1` (pushed)
- **Release:** v1.5.1 (PATCH — single self-update bug fix)
- **Files changed:** 4 (0 new smaqit source, 4 modified)
- **Tasks planned:** 1 (Task 087)
- **Tasks closed:** 1 (Task 084, documentation finalization)
