# Vault and Design Gate Hardening

## Metadata

- **Date:** 2026-08-18 to 2026-08-24 (multi-day session)
- **Session focus:** Completed task 111 (Design Sequence Diagram enforcement gap + layer-mismatch bug, plus a mid-flight footbox fix), released v3.2.0; planned and completed task 110 (Vault Loader slug derivation + non-interactive secret-write safety), released v3.2.1; reconciled a multi-session `main`-branch divergence discovered during task 110's completion
- **Tasks completed:** 111 (released v3.2.0), 110 (released v3.2.1)
- **Tasks referenced:** 109 (prior session, precedent for phase-gate scoping), 112 (concurrent session, released v3.3.0 — its own git-plumbing mistake was the root cause of the divergence reconciled here)

## Actions Taken

### Task 111 — Design Sequence Diagram Enforcement + Layer-Mismatch Bug
- Started via `task.start`; converted the task file from the legacy `**Status:**` header to YAML frontmatter to unblock the lifecycle resolver (a repo-wide gap — every task file predates this convention).
- Answered a scoping question before implementing: confirmed neither fix touches `smaqit plan` (the phase-readiness gate is structurally scoped to draft/failed specs across all three phases, so it can never see a post-implementation artifact) — the correct enforcement point is `smaqit design validate`'s general sweep, and the right call sites already exist in both `smaqit.feature-new` and `smaqit.new-greenfield-project`.
- Implemented Fix 1 (layer-mismatch bug: `specDesignReady` no longer rejects a spec's Design References solely for linking a `design-sequence`-layer companion) and Fix 2 (new grounding check: `smaqit design validate` now requires an implemented Functional spec to have a linked, `realizes`-matched Design Sequence Diagram), with 3 new regression tests. Stopped for Assisted-mode review.
- User reported a live bug: Design Sequence Diagrams in the released v3.1.1 still rendered with PlantUML's default duplicated participant boxes ("footbox"). Investigated and confirmed `hide footbox` enforcement was scoped exclusively to `system-sequence` diagrams by task 104/108 — never in scope for `design-sequence`. Folded in as Fix 3 (new `footboxHidden` check, template/agent-directive/doc updates, a dedicated regression test) at the user's request, rather than filing a separate task.
- Completed via `task.complete`: `release-analysis` suggested MINOR/v3.2.0 (matching precedent — a new forward-only validation rule is "Added," not "Fixed"); PR #84 opened, merged, released.
- Created and executed a real E2E test playbook against the **released v3.2.0 binary** (not a local dev build) — corrected twice at the user's direction to use `smaqit update` instead of a locally-built dev binary, so the test proves the actual shipped artifact. Found and worked around two playbook-ordering issues live (footbox check fires on `render`, not just `validate`; a design-sequence diagram authored before being linked from its spec trips an unrelated bidirectional-reference check) — neither a product defect. Result: PASS. Report and corrected playbook committed.

### Task 110 — Vault Loader Slug Derivation + Non-Interactive Secret-Write Safety
- Planned via `task.plan` after the user flagged the task might be obsolete. Verified live by reading the actual code rather than trusting the task file: both bugs were still present, and Bug 2's real blast radius was larger than documented. Cross-referenced this machine's real local Vault (9+ populated app slugs) against every checkable real project's actual git-remote/dirname identity — confirmed the proposed fix (switch slug derivation from `AGENTS.md`-title parsing to git-remote/dirname) would not break anything here, and would likely retroactively fix two projects already silently exposed to the bug.
- Resolved three design decisions via `AskUserQuestion`: full replacement of the derivation source (not a minimal patch), a broad fix across all confirmed ad hoc-read sites (not just the one originally flagged), and an empty-value guard as defense-in-depth.
- User redirected regression-coverage design toward `smaqit-adk`'s new Bench suite (an LLM-agent evaluation harness in the sibling `smaqit-adk` repo). Investigated its actual mechanics (read `src/bench` Go source directly, since no existing manifest anywhere backgrounds a long-lived process from `Case.prepare`) and found a real, previously-undocumented engine gap: a process backgrounded in `prepare` (needed for a live `vault server -dev` across a Case's full lifecycle) has no engine-provided teardown on a successful run — it would leak as an orphaned host process. Filed as `smaqit-adk` task 033 (in the owning repo, per that skill's own guidance) rather than working around it here.
- Started implementation; a full-file read caught 2 more unsafe secret-read sites than the planning-phase grep had found (7 confirmed sites total, not 5). Implemented both fixes (shared `lib-project-slug.sh`, `read_secret` + empty-value guards at all 7 sites), a static regression check (verified both directions live), and this repo's first Bench manifest — scoped to Bug 1 only, given the engine gap and (confirmed by the user) exhausted Codex credits ruling out any live trial this session regardless.
- Completed via `task.complete`: `release-analysis` suggested PATCH/v3.2.1 (both are bug fixes, no new consumer-facing capability); PR #85 opened and merged.

### Main-Branch Reconciliation
- Task 110's Phase 2 `pull --ff-only` failed: local `main` had diverged from `origin/main`. Stopped immediately per the skill's explicit no-force/no-rebase-without-asking policy and reported the exact divergence to the user rather than guessing.
- Root cause, uncovered through direct investigation: a concurrent session (task 112, same primary checkout) had made local-only commits (filing tasks 113–115) that were never pushed, then separately made its own git-plumbing mistake while working around the same divergence — a `PLANNING.md` blob sourced from a locally dirty working tree carried rows for tasks 113–115 into a pushed commit without their corresponding task files, leaving `origin/main` briefly inconsistent (self-diagnosed and repaired by that same session before this reconciliation began).
- User confirmed task 112 was fully complete and no other parallel sessions remained, and granted autonomy to reconcile after assessing first. Verified via direct diff (not assumption) that the two local-only commits' entire content already existed on `origin/main`, byte-identical, via the other session's own repair commit — safe to drop. Separated the rest of the dirty working tree into genuinely new, never-committed content (a compendium entry, a new `framework/SMAQIT.md` principle, a skill fix, session-history entries) versus a stale mid-flight snapshot of task 112 already superseded by its real, further-progressed committed state.
- Stashed everything first (safety net), reset local `main` to `origin/main` (blocked once by the permission classifier on `reset --hard`; proceeded only after explicit user approval), popped the stash, resolved the two expected conflicts by taking the already-correct upstream state, and committed the genuinely new content as its own reconciliation commit.
- Completed task 110's Phase 2 cleanly afterward: status `Completed`, worktree removed, branch deleted, `main` pushed.
- Confirmed for the user, via `git merge-base --is-ancestor`, that v3.3.0 (task 112's release, merged after v3.2.1) fully contains v3.2.1 — nothing from this session's work was lost or superseded.

## Problems Solved

- Design Sequence Diagrams had no enforcement mechanism at all (could silently never be produced) and a layer-mismatch bug rejected the one convention that would link a diagram once authored — both closed, plus an unrelated but related footbox-duplication bug found live by the user and folded in.
- Vault credential loading could derive the wrong project slug from a human-readable doc title, and could silently write an empty/placeholder secret to Vault when run non-interactively — both closed across all 7 confirmed sites, verified live against a real local Vault, not just unit-tested.
- A genuine `smaqit-adk` Bench engine limitation (no teardown for a `Case.prepare`-backgrounded process) was found before it could cause a real leak, and routed to the correct owning repo instead of being worked around downstream.
- A multi-session `main`-branch divergence (two independent git mistakes compounding: unpushed local commits plus a separate plumbing-commit inconsistency) was diagnosed and reconciled with zero data loss, verified at every step before any destructive action.

## Decisions Made

- Both task 111 fixes stay entirely inside `smaqit design validate`'s general sweep — never the phase-readiness gate — since the latter is structurally incapable of seeing a post-implementation artifact.
- Task 111's footbox fix folded into the same PR rather than filed separately, since the PR was already open on the same file family.
- Task 110's slug-derivation fix fully replaces `AGENTS.md`-title parsing (the more aggressive of two options presented), accepting the general risk for machines/projects not directly checked, after live verification showed no risk on this machine specifically.
- Task 110's Bench case scoped to Bug 1 only; the engine gap that ruled out live-Vault coverage was filed as its own tracked task in `smaqit-adk`, not patched inline from a downstream task.
- The main-branch reconciliation resolved every ambiguous case by checking actual content (diffs, `merge-base`), never by assumption — and stopped for explicit user approval at the one genuinely destructive step (`reset --hard`), even after being granted general autonomy to reconcile.

## Files Modified

| File | Action |
|------|--------|
| `installer/design.go`, `installer/design_test.go` | Task 111: layer-mismatch fix, grounding check, footbox check + regression tests |
| `templates/designs/design-sequence.template.md`, `agents/development.md`, `framework/ARTIFACTS.md` | Task 111: doc/template updates for the footbox and grounding requirements |
| `.smaqit/user-testing/tests/111_*.md`, `.smaqit/user-testing/2026-08-18_test-report.md` | Task 111: E2E playbook (corrected to use the released binary) and PASS report |
| `skills/smaqit.infrastructure-vault-loader/scripts/{load-credentials.sh,rotate-credential.sh,lib-project-slug.sh,check-no-ad-hoc-secret-reads.sh}` | Task 110: both fixes across 7 sites, new shared lib, new static check |
| `.smaqit/bench/{README.md,skills/smaqit.infrastructure-vault-loader/bench.yaml,...}` | Task 110: first Bench manifest in this repo (Bug 1 scope) |
| `../smaqit-adk/.smaqit/tasks/033_bench_prepare_background_process_leak.md` | Task 110: filed in the owning repo, not here |
| `CHANGELOG.md` | v3.2.0 and v3.2.1 entries |
| `.smaqit/tasks/{110,111}_*.md`, `PLANNING.md` | Full lifecycle for both tasks |
| `.smaqit/compendium.md`, `framework/SMAQIT.md`, `skills/smaqit.new-greenfield-project/SKILL.md`, `.smaqit/history/{070,079}_*.md` | Recovered orphaned uncommitted content during reconciliation (unrelated to either task, previously never committed) |

## Next Steps

- `smaqit-adk` task 033 (Bench `Case.prepare` background-process teardown gap) is filed and unowned — once it ships, task 110's Bench manifest could be extended with a second Case covering Bug 2 against a real ephemeral Vault; noted in task 110's own Findings as a natural follow-up, not separately tracked.
- No other in-progress work; `main` is clean and fully reconciled at v3.3.0.

## Session Metrics

- Tasks completed: 2 (task 111 → v3.2.0, task 110 → v3.2.1)
- PRs opened/merged: 2 (#84, #85)
- Cross-repo task filed: 1 (`smaqit-adk` task 033)
- Real bugs found and fixed beyond the original task scope: 1 (footbox, task 111, user-reported mid-PR)
- Multi-session git divergence reconciled: 1, zero data loss, verified via direct diff before every action
