# K3s Skill Retirement Gap

## Metadata

- **Date:** 2026-09-12
- **Session focus:** Planned, implemented, and shipped task 119 (v3.5.1) fixing three live-verified defects in the k3s onboarding/deployment skill family from tasks 116/117/118. Post-release, discovered and live-verified a real, previously undocumented installer gap: `smaqit update` never prunes a skill directory that drops out of a release's embed, contradicting `.smaqit/compendium.md`'s own claim of a "tombstone" mechanism that does not actually exist in code.
- **Tasks completed:** 119 (released v3.5.1, PR #89)
- **Tasks referenced:** 116, 117, 118 (the k3s onboarding/deployment family task 119 fixes follow-up defects in)

## Actions Taken

### Task 119 — Fix Registry-Write Semantics, Deployment Sizing, and Lifecycle Gaps
- Verified all four of the task file's findings directly against current code before planning, rather than trusting the task text at face value: confirmed Findings 1 (registry full-overwrite), 2 (oversized deployment defaults), and 4 (orphaned `smaqit.infrastructure-onboard-k3s-app` stub, zero callers) as real; determined Finding 3's premise was wrong for this repo — `smaqit.task-start`/`smaqit.task-complete` are owned by the sibling `smaqit-extensions` repo, confirmed by memory and a direct directory check.
- Ran `smaqit.task-plan` (Mode B): scored Complex (task file had no Implementation Steps/Design Decisions sections), did direct-read discovery instead of spawning Explore subagents (small, already-identified file set), and produced a plan the user approved with one correction — drop Finding 3 from scope entirely, per explicit user direction.
- Updated the task file (Design Decisions, Implementation Steps, Files to Create/Modify, trimmed Acceptance Criteria, Finding 3 moved to Out of Scope) before starting.
- `smaqit.task-start`: created the task branch/worktree; hit and fixed two pre-existing task-file hygiene gaps surfaced by the pipeline — a missing `## Notes` section blocking `task-context.sh`'s legacy-format extraction, and a real `CONTRIBUTING.md` violation (the task file named a downstream project, "Magnificah," by its real name in prose) — corrected both before implementation. Issue triage returned Historical only (one tangentially related closed `cli/cli` issue about a different command's destructive overwrite; no blocking findings).
- Implemented all three in-scope findings: rewrote `smaqit.infrastructure-request-k3s-onboarding`'s Step 3 to fetch-then-append instead of blind-overwrite (verified with a local bare-git simulation confirming an existing unrelated entry survives); tokenized `deployment.yaml.template`'s `replicas`/CPU/memory request defaults (`__REPLICAS__`/`__CPU_REQUEST__`/`__MEM_REQUEST__`, defaulting to `1`/`50m`/`32Mi`, verified via a rendered-and-linted sample manifest); deleted the orphaned `smaqit.infrastructure-onboard-k3s-app` directory and cleaned up two now-stale mentions of it elsewhere. Updated `installer/main_test.go`'s skill-count assertions (30→29) and `.smaqit/compendium.md`'s onboarding-family entry.
- `smaqit.task-complete` Phase 1: computed v3.5.1 (PATCH) via `release-analysis`, correctly reasoning through a tricky boundary situation — `origin/main`'s trunk had several untagged administrative commits ahead of the v3.5.0 tag from the prior session (task 117 bookkeeping, a dogfood `smaqit init`/`update` sync) that were excluded from the delta as already-documented or non-product housekeeping, not task 119's own unreleased work. Opened PR #89, then caught and corrected an own mistake: the PR body had included the "🤖 Generated with Claude Code" footer, which this repo's `CLAUDE.md` explicitly forbids — fixed via the GitHub REST API after `gh pr edit` hit an unrelated GraphQL Projects-classic deprecation error.
- Hit a CHANGELOG rebase conflict while promoting the pending entry onto the PR branch — caused by the branch's own implementation commit having independently added raw (unannotated) versions of the same three entries. Rather than resolve mid-conflict, aborted the rebase, stripped the redundant hunk out of the implementation commit via amend, and re-ran the rebase cleanly.
- `smaqit.task-complete` Phase 2 (user said "merged"): confirmed via `gh pr view`, merged `origin/main` into local `main` cleanly, marked task 119 Completed, removed the worktree, force-deleted the local branch, rebuilt the workspace.

### Post-release: k3s skill retirement gap discovery
- User updated this machine to v3.5.1 and asked for confirmation the fixes landed globally. Direct file checks (not assumption) showed a mixed result: the registry-append and sizing-token fixes landed correctly in both `~/.claude/skills/` and `~/.agents/skills/` (fresh mtimes), but the orphaned `smaqit.infrastructure-onboard-k3s-app` directory was **still present** in both, with a stale mtime predating the update.
- Traced the root cause in `installer/main.go`: `cmdInstallGlobal()` → `copyEmbeddedDir()` only walks the *current* binary's embedded tree and writes/overwrites files — it never compares against what's already on disk to prune anything absent from the new embed. Grepped the entire installer source (`main.go`, `update.go`, `design.go`, `spec.go`) for `tombstone`/`retired`/`legacy`-style cleanup logic and found none.
- This directly contradicts `.smaqit/compendium.md`'s own "How must a release retire a previously shipped skill from existing projects?" entry, which describes a "persistent installer tombstone" mechanism that does not exist anywhere in code — either aspirational documentation never implemented, or a silent regression. Reported this plainly rather than assuming the compendium was accurate.
- User declined filing a task for the underlying installer fix (for now) and asked instead for a manual one-machine cleanup. Verified no `CLAUDE_CONFIG_DIR`/shared-skills override was in play, confirmed both stale directories contained exactly one file (`SKILL.md`, no unexpected nested content) with a pre-update mtime, then deleted both and confirmed removal.
- Produced a self-contained, safety-gated prompt for the user to paste into another machine's session to perform the same verify-then-clean flow (version check, directory-contents check, mtime comparison against a sibling fixed file, explicit flag-instead-of-delete conditions).

## Problems Solved

- Three live-verified defects in the shipped k3s onboarding/deployment skill family (silent registry deregistration risk, quota-exceeding default sizing, an orphaned/dead skill stub) closed via task 119, released as v3.5.1.
- Two pre-existing task-file hygiene gaps caught and fixed before they could propagate: a missing `## Notes` section blocking the research-map/triage pipeline for legacy-format task files, and a real downstream-project-naming violation of this repo's own `CONTRIBUTING.md` convention.
- An own mid-flow mistake (the forbidden AI-authorship PR footer) caught and corrected before merge, rather than left to ship.
- A self-inflicted CHANGELOG rebase conflict resolved by restructuring the commit history rather than manually resolving conflict markers mid-rebase, consistent with this project's "never auto-resolve a rebase conflict" standing rule.
- A real, previously undocumented (and contradicted-by-its-own-docs) installer gap discovered via direct live verification rather than assumption: retired skills are never pruned from already-installed global directories by `smaqit update`, on any machine, for any past or future skill retirement.

## Decisions Made

- Finding 3 (a `smaqit.task-start`/`smaqit.task-complete` branch-protection-to-chore-PR fallback) dropped entirely from task 119's scope, per explicit user direction — not rescoped to a companion task in `smaqit-extensions` this session.
- Finding 4 resolved as outright removal of `smaqit.infrastructure-onboard-k3s-app`, not a fold — its content was already fully covered by `smaqit.infrastructure-request-k3s-onboarding`'s own Pre-conditions/Gotchas.
- The optional `manifest-lint.sh` quota-aware warning suggested alongside Finding 2 was explicitly left out of scope — no quota-ceiling input exists anywhere in the linter's current interface, and building one wasn't a hard requirement of any Acceptance Criterion.
- The newly discovered skill-retirement pruning gap was **not** filed as a task this session, per explicit user direction — the user chose an immediate manual per-machine cleanup instead, with a follow-up prompt prepared for other machines.

## Files Modified

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` | Fetch-then-append registry-write rewrite; cleaned up stale references to the removed onboarding stub |
| `skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template` | Tokenized `replicas`/CPU/memory request defaults |
| `skills/smaqit.infrastructure-onboard-k3s-app/` | Deleted — orphaned, superseded |
| `installer/main_test.go` | Skill-count assertions 30 → 29 |
| `.smaqit/compendium.md` | Onboarding-family Q&A entry updated |
| `CHANGELOG.md` | `v3.5.1` entry |
| `.smaqit/tasks/119_*.md`, `PLANNING.md` | Full lifecycle for task 119 |
| `~/.claude/skills/smaqit.infrastructure-onboard-k3s-app/`, `~/.agents/skills/smaqit.infrastructure-onboard-k3s-app/` | Deleted manually on this machine (outside the repo) — installer has no automated pruning for this |

## Next Steps

- The installer's skill-retirement pruning gap remains unfixed at the code level — `cmdInstallGlobal`/`copyEmbeddedDir` still has no mechanism to remove a directory absent from the current embed. Every past skill retirement in this framework's history likely left the same kind of stale directory on any machine that updated across that boundary; not filed as a task this session per explicit user direction, but flagged clearly and left for a future decision.
- The self-cleanup prompt was handed to the user for other machines; not run or verified elsewhere this session.
- Live verification against a real platform repo (an actual PR open/merge cycle, a real registry file) for task 119's Finding 1 fix remains outstanding — same class of environment limitation already recorded as Follow-up in tasks 117/118, not newly introduced.
- `.smaqit/compendium.md`'s "How must a release retire a previously shipped skill from existing projects?" entry describes a mechanism that doesn't exist in code — worth correcting or removing in a future pass so it stops asserting something untrue.

## Session Metrics

- Tasks completed: 1 (119 → v3.5.1, PR #89)
- PRs opened/merged: 1 (#89)
- Skills removed: 1 (`smaqit.infrastructure-onboard-k3s-app`, source); manually cleaned from 2 global directories on this machine
- Real defects fixed: 3 (registry overwrite risk, unsafe default sizing, orphaned skill stub)
- Own mistakes caught and corrected mid-session: 2 (forbidden PR footer; self-inflicted CHANGELOG rebase conflict)
- New installer gap discovered via live verification (not yet fixed): 1 (no skill-retirement pruning on `smaqit update`)
