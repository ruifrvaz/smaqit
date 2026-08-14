# Design Gate Scope Release

**Date:** 2026-08-14
**Session focus:** Implement, ship, and release task 109 (phase design-readiness gate scoping bug), then clean up downstream-project naming spillage in task files.
**Tasks completed:** 109 (Phase Design-Readiness Gate Scans All Active Specs, Not Just the Touched Feature) — released as v3.1.1

## Actions taken

- Started task 109 via `task.start`: created its branch/worktree, ran the project-research task-only refresh, ran issue triage (no third-party tools — clear), and set the task to `In Progress`/Assisted.
- Implemented the fix in `installer/spec.go`: added a shared `isCycleRelevant(status)` predicate (`draft`/`failed`) reused by `getPhaseDesignGateSpecs` instead of its old "every non-deprecated spec" sweep, and changed `validatePhaseDesignReadiness` to aggregate every blocking spec into one error instead of failing on the first.
- Explicitly deferred a related but distinct idea (a per-layer design-pair exemption flag, e.g. for projects that never give Coverage specs a design pair) as its own future follow-up rather than folding it into this fix — recorded the reasoning in the task's Design Decisions.
- Added/updated `installer/spec_test.go` regression coverage: `TestGetPhaseDesignGateSpecsScopesToCurrentCycle` (the task's own reported scenario — a legacy `implemented` spec no longer blocks an unrelated feature's plan) and `TestValidatePhaseDesignReadinessAggregatesAllFailures`; updated the existing layer-scope test to use draft-status fixtures. Full `go build`/`go vet`/`go test ./...` passed after `make prepare` regenerated the gitignored embedded-asset staging.
- Walked task 109 through its full Assisted-mode, PR-gated completion lifecycle: Phase 1 (`release-analysis` in Task mode suggested `v3.1.1`/PATCH; opened PR #83 titled `Prepare release v3.1.1`; wrote and then promoted the pending `CHANGELOG.md` entry), user confirmed the merge, Phase 2 (verified `MERGED` via `gh pr view`, pulled `main`, marked the task `Completed`, removed the worktree, force-deleted the local branch, rebuilt the workspace file). `post-merge-release.yml` tagged and published `v3.1.1` automatically.
- Scrubbed the downstream project name (`iodis-crm-poc`) out of task 109 and task 110's files per `CONTRIBUTING.md`'s naming rule — genericized task 109's mentions, and replaced task 110's (where the real name was load-bearing evidence for a slug-truncation bug) with an illustrative fictional project name that preserves the same bug mechanics.
- Investigated filing the Phase-2 lifecycle-resolver bug (`9_resolve_task_lifecycle.sh` rejected `Status: PR Open` for `--purpose complete`, hit live during task 109's own Phase 2) into the sibling `smaqit-extensions` repo per the user's request — found it was already discovered and fixed there same-day (commit `d50e4b9`, part of that repo's own task 027), so no duplicate task was filed.

## Problems solved

- The phase design-readiness gate (`smaqit plan --phase=develop|deploy|validate`) blocked unrelated feature work in any project incrementally adopting the PlantUML design-pair convention, because it scanned every active spec in a phase's layers instead of only the current cycle's. Root-caused live in a downstream project by two independent sessions the same day; fixed and released as v3.1.1 this session.
- Encountered the already-fixed `9_resolve_task_lifecycle.sh` Phase-2 gap directly while completing task 109 — worked around it manually (independently re-verifying every fact the resolver would have returned) rather than blocking completion on a sibling-repo tooling issue.

## Decisions made

- Scope the design gate to `draft`/`failed` status (reusing the same predicate `filterSpecsByStatus` already used for pending-count accounting) rather than inventing a new concept.
- Switch `validatePhaseDesignReadiness` to aggregate reporting now that its scoped input is expected to be small — removes the fix-one/rerun/find-the-next loop.
- Defer the per-layer design-pair exemption question (e.g. "Coverage never needs a design pair") as a separate follow-up task rather than bundling a policy decision into a structural scoping bug fix.
- Did not rewrite two already-pushed `main` commit messages that still literally name the downstream project (`d977777`, `1bd1f22`) — CONTRIBUTING.md's naming rule covers file content, not commit messages, and fixing it would require a destructive history rewrite of shared `main`; left for the user to decide.

## Files modified

- `installer/spec.go` — `isCycleRelevant`, scoped `getPhaseDesignGateSpecs`, aggregate `validatePhaseDesignReadiness`
- `installer/spec_test.go` — new/updated regression tests
- `CHANGELOG.md` — `v3.1.1` entry (pending → promoted)
- `.smaqit/tasks/109_phase_design_gate_scope_not_feature_bound.md` — full lifecycle (Not Started → In Progress → PR Open → Completed), Design Decisions, Findings, downstream-name scrub
- `.smaqit/tasks/110_vault_loader_slug_derivation_and_silent_placeholder_write.md` — downstream-name scrub only (task itself not started)
- `.smaqit/tasks/PLANNING.md` — task 109 moved Active → Completed
- `.smaqit/references/project-research.md` — task 109 block added
- `smaqit.code-workspace` — rebuilt after worktree cleanup

## Next steps

- Task 110 (Vault Loader slug derivation + silent placeholder-write bug) is next up in Active, Medium priority, not yet started.
- User may want the two `main` commit messages (`d977777`, `1bd1f22`) that still name the downstream project rewritten — flagged, not actioned, needs an explicit decision given it requires a `main` history rewrite.
- An unrelated, still-uncommitted local edit to `CLAUDE.md` exists on `main` (adds a "no AI-authorship PR footer" instruction) — pre-dates and is unrelated to this session's work; left untouched per instruction.

## Session Metrics

- Tasks completed: 1 (task 109, released as v3.1.1)
- Files modified: 8 (see above)
- Commits pushed to `main` this session: 8 (task start, pending changelog, PR-open state, downstream-name scrub, task-complete state, workspace rebuild, plus the PR's own 2 commits merged in)
- PR: #83, merged, tagged and released as `v3.1.1`
