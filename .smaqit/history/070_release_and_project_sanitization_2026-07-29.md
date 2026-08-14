# Release and Project Sanitization

**Date:** 2026-07-29
**Session Focus:** Cut release v1.11.0 (task 096 — existing-unmanaged provisioning mode), then perform a repo-wide audit and redaction of downstream project names that had bled into smaqit's own task and history files.
**Tasks Referenced:** 096, 097
**Tasks Completed:** None (task 096 was already complete before this session; task 097 was executed directly without formal task lifecycle)

---

## Actions Taken

### Release v1.11.0
- Ran `smaqit.release-analysis`: identified one PR since v1.10.0 — task 096 (existing-unmanaged provisioning mode). Assessed as MINOR (new feature, backward-compatible). Suggested v1.11.0.
- Obtained user approval and prepared files: promoted `[Unreleased]` to `[1.11.0]` in `CHANGELOG.md` with 1 Added, 4 Changed, and 1 Fixed entry covering the new provisioning mode and its reviewed `load-credentials.sh` fix. Bumped `installer/main.go` Version from `1.10.0` to `1.11.0`. Build verified.
- Committed (`e9edc32`), created annotated tag `v1.11.0`, pushed both to origin via GNOME Keyring SSH agent (`/run/user/1000/gcr/ssh`). Tag push confirmed on remote.

### Project Bleed Cleanup
- Performed a comprehensive repo-wide audit for downstream project names: one project had 19 instances across 4 files, another had 1 instance in 1 file. The remaining names checked were already clean.
- Redacted all instances in `.smaqit/tasks/092_*.md` (5), `.smaqit/tasks/094_*.md` (7), and `.smaqit/tasks/095_*.md` (6), replacing with anonymized phrasing ("a downstream project", "that downstream project", "the downstream project that surfaced this").
- Deleted `.smaqit/tasks/097_redact_downstream_project_names.md` — the cleanup task superseded by this session's work.
- Updated `.smaqit/tasks/PLANNING.md` — removed task 097 from the Active table.
- Updated `CONTRIBUTING.md` — added a new rule under Scope and Style: *"Never name downstream (consumer) projects in task files, history files, CHANGELOG entries, or shipped skill documentation."* The sole exception is `mario-hello`, smaqit's own test fixture.
- Final audit confirmed zero remaining matches for any downstream project name across all `.md` files.

## Problems Solved

- **Release v1.11.0 shipped** — task 096's existing-unmanaged provisioning mode is now published, with a clean changelog entry and verified build.
- **19 instances of project bleed cleaned** — a downstream project name redacted from committed and untracked task files without losing the technical context those references provided (the actual gap descriptions, task IDs on the downstream project, and dates are retained; only the project name/handle itself is anonymized).
- **Convention encoded** — the "never name downstream projects" rule is now explicit in `CONTRIBUTING.md`, so future task authors have a durable reference rather than relying on tribal knowledge or post-hoc cleanups.

## Decisions Made

- **Task 097 deleted rather than completed** — since the cleanup was performed directly in this session without the formal `task.start`/`task.complete` lifecycle, and the task's own acceptance criteria (repo-wide audit, scope decision, convention documentation) were all satisfied, the task file was removed rather than preserved as a record of a purely-cleanup task.
- **History 066's pre-existing cleanup left as-is** — the working tree had already removed the project-name enumeration line from history 066 before this session began; this change was left as a pre-existing uncommitted modification rather than being reverted or re-applied.

## Files Modified

### Release v1.11.0
- `CHANGELOG.md` — v1.11.0 section added
- `installer/main.go` — Version `1.10.0` → `1.11.0`

### Project Bleed Cleanup
- `.smaqit/tasks/092_feature_deploy_stale_push_trigger_assumption.md` — 5 instances redacted
- `.smaqit/tasks/094_feature_new_no_e2e_browser_gate.md` — 7 instances redacted
- `.smaqit/tasks/095_feature_new_per_phase_worktree_spawn.md` — 6 instances redacted
- `.smaqit/tasks/PLANNING.md` — removed task 097 row
- `CONTRIBUTING.md` — added project-name convention

### Deleted
- `.smaqit/tasks/097_redact_downstream_project_names.md` — superseded

## Next Steps

- Remaining active tasks: 095 (per-phase worktree spawn), 094 (E2E browser gate), 077 (retroactive specs), 074 (extensible through templates), 071 (Q&A agent), 070 (E2E boundary enforcement)
- History 066 and compendium/PLANNING pre-existing modifications remain uncommitted in the working tree
- `smaqit-extensions` reinit fix (from user next-steps) is a cross-repo item

## Session Metrics

- **Release:** v1.11.0 (published, `post-merge-release` triggered)
- **Instances redacted:** 19 + 1 (two downstream project names)
- **Files modified:** 8 | **Deleted:** 1
- **Commit:** `e9edc32` — pushed to `main` + annotated tag `v1.11.0`
- **Audit:** Zero project-name leaks confirmed across entire repo
