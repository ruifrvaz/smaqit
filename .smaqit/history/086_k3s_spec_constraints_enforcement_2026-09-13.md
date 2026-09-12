# K3s Spec Constraints Enforcement

## Metadata

- **Date:** 2026-09-13
- **Session focus:** Task 120 — surface and fail gracefully on missing `existing-k3s` Infrastructure-spec fields
- **Tasks completed:** 120 (released v3.5.2, PR #90)

## Actions Taken

- Ran `session.start`, loading project state, memory, `PLANNING.md`, and the compendium. Identified
  task 120 as the top-priority, fully-scoped, ready-to-start item.
- Started task 120 (`smaqit.task-start`): created its owner branch/worktree, ran the research-map
  task-only refresh (task block written, reusing existing Kubernetes/cert-manager references —
  no new tools), and ran issue triage, which correctly exited as not-applicable (the task only
  edits smaqit's own documentation prose, no new script/API/library integration).
- Implemented the fail-fast fix across four files:
  - `templates/specs/infrastructure.template.md` — added five new conditional `## Constraints`
    rows (`Registry File Path`, `Machine Slug`, `Ingress Class`, `ClusterIssuer`,
    `Namespace Quota`) alongside the pre-existing `Platform Repo` row.
  - `agents/infrastructure.md` — added a MUST rule directing the Infrastructure spec-writing agent
    to populate those six rows for an `existing-k3s` target, sourced only from the platform team's
    own documentation, with an explicit fallback to flag as an Untestable-Criteria gap rather than
    guess.
  - `skills/smaqit.input-deployment/SKILL.md` — added a mode-resolution-time note to the
    `existing-k3s` value description telling the operator these fields must already exist in the
    spec.
  - `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` — sharpened Pre-conditions to
    name the exact Constraints rows and added the missing Failure Handling row for a spec that
    doesn't declare them.
  - Checked `smaqit.feature-new` for the same silent assumption — confirmed it does not duplicate
    it (never calls the onboarding-request skill; only reads an already-populated kubeconfig from
    Vault for a post-MVP redeploy).
- Completed task 120 (`smaqit.task-complete`): Phase 1 committed implementation, computed the
  release version via `smaqit.release-analysis` (Task mode, PATCH severity), opened PR #90
  ("Prepare release v3.5.2"), wrote and promoted the CHANGELOG entry. Assisted mode stopped for
  review; user confirmed the merge, and Phase 2 verified the merge via `gh pr view`, merged
  `origin/main` into local `main`, marked the task Completed, and cleaned up the worktree/branch.
  Post-merge automation tagged and released v3.5.2 automatically.

## Problems Solved

- Fixed a mistaken inclusion of the "🤖 Generated with Claude Code" footer in PR #90's description,
  which this repo's `CLAUDE.md` explicitly forbids — caught and corrected immediately via the
  GitHub REST API after `gh pr edit`'s GraphQL call failed on an unrelated deprecated-field error
  in this environment's older `gh` CLI version.
- Confirmed a passing concern about "local-only bookkeeping commits" inflating the PR diff was
  unfounded: origin/main had already received several housekeeping commits (task 119 Phase 2
  cleanup, a version-string sync, session history) ahead of the v3.5.1 tag, so the actual merge
  base between origin/main and the task branch was already past them — the PR diff was correctly
  minimal (task 120's own 9 files only).

## Decisions Made

- Extended the template's Constraints table with rows for all six fields named across the task's
  Description (the onboarding triplet — Platform Repo/Registry File Path/Machine Slug — plus the
  deploy-time trio — Ingress Class/ClusterIssuer/Namespace Quota) rather than just the three named
  literally in Acceptance Criterion #1, since the Design Decisions' intent ("name these fields...
  the same way it would for any other provisioning-mode-specific fact") applied equally to both
  groups and the template's Constraints table was the natural single home for all of them.
- Deliberately left `smaqit.infrastructure-deploy-k3s-app` unedited even though its own
  Pre-conditions generically assume these fields exist — it was never one of the task's three named
  touch points or the `smaqit.feature-new` check; fixing the upstream source should make its
  existing generic assumption correct without a matching edit there. Recorded as Follow-up.

## Files Modified

- `templates/specs/infrastructure.template.md` — five new conditional Constraints rows
- `agents/infrastructure.md` — MUST rule naming required `existing-k3s` spec fields
- `skills/smaqit.input-deployment/SKILL.md` — `existing-k3s` description now names the requirement
- `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` — Pre-conditions + Failure
  Handling row
- `CHANGELOG.md` — v3.5.2 entry
- `.smaqit/tasks/120_surface_and_fail_gracefully_on_missing_existing_k3s_spec_fields.md`,
  `.smaqit/tasks/PLANNING.md` — task lifecycle bookkeeping
- `.smaqit/references/project-research.md` — task 120 research block
- `smaqit.code-workspace` — worktree lifecycle bookkeeping

## Next Steps

- `smaqit.infrastructure-deploy-k3s-app`'s own Pre-conditions/Failure Handling still reference
  these fields only generically rather than by the exact Constraints-table row names this task
  introduced — worth tightening once the new field names are load-bearing in real specs.
- The installer skill-retirement pruning gap (`cmdInstallGlobal`/`copyEmbeddedDir` in
  `installer/main.go` never deletes a directory absent from the current embed) remains unfixed and
  un-filed as a task, carried over from session 085.
- `PLANNING.md`'s remaining Active candidates: task 113 (self-contained, no blockers) is the most
  ready pick; task 114 needs a scoping conversation first (and 115 likely depends on it); tasks
  071/074 may already be satisfied by current code and worth a reconciliation pass; task 115 still
  names a retired pre-k3s tenancy model and likely needs re-scoping.

## Session Metrics

- Tasks completed: 1 (task 120)
- Release shipped: v3.5.2 (PR #90)
- Files modified: 8 (4 source files + CHANGELOG + 2 task-bookkeeping files + workspace file)
- PRs opened/merged: 1
