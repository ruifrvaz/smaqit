# `smaqit.feature-new` — Per-Phase `task.start`/`task.complete` Spawns Redundant Branches/Worktrees Instead of One Shared Feature Branch

**Status:** Completed
**Created:** 2026-07-27
**Completed:** 2026-08-05

## Description

`smaqit.feature-new`'s own SKILL.md instructs, for each of Phases 1, 2, 3, and 4, independently: *"Invoke `smaqit.task-start` for the Phase N task"* (lines 43, 57, 70, 108) and *"Invoke `smaqit.task-complete` for the Phase N task"* (lines 53, 62, 104, 113). Per `smaqit-extensions`' own `smaqit.task-start`/`smaqit.task-complete` contract (built by `smaqit-extensions` task 018, "Converge Worktree-Aware Task Lifecycle" — see that task's Acceptance Criteria: *"Starting a task creates or reuses its sibling worktree"* / *"`task.complete [id]` merges the task branch, removes its worktree, safely deletes the merged branch"*), this is unconditional, uniform behavior for **any** task ID — it has no concept of "this task is phase N of an existing feature cycle, reuse phase 1's branch."

Taken literally, a `smaqit.feature-new` run would create and merge **four separate branches/worktrees** (one per phase 1–4), each independently merged into `main` at its own `task.complete`. This directly contradicts Phase 3 (Deployment)'s own text one section later, which describes the Deployment agent as owning *"commit and push **the feature branch**, create **the PR**... pause for human merge"* (line 99, singular) — i.e. Phase 3's own wording assumes one shared feature branch survives from Phase 1 through Phase 3, not four independent ones. The skill is self-contradictory as written: it cannot simultaneously mean "each phase gets its own task-start/task-complete cycle" (literal Phase 1/2/3/4 instructions) and "there is one feature branch that Phase 3 pushes and PRs" (Phase 3's own prose).

In practice, every real `smaqit.feature-new` run so far (tasks 005/008–012, 016/017–021 on a downstream project) has produced exactly one PR at the Deployment phase, meaning the executing agent each time silently deviated from a literal per-phase `task.start`/`task.complete` reading to do the sensible thing — sharing one branch/worktree across all five phase tasks, invoking the real git merge only once. This is tribal knowledge, not something either this skill or `smaqit-extensions`' `smaqit.task-start`/`task.complete` documents or enforces; it has already caused one real mistake caught by a user mid-session (task 016 on that downstream project, where a session briefly worked in the wrong checkout), and was independently rediscovered as a live concern (before any mistake occurred this time) during that downstream project's task 022 planning, 2026-07-27 — the user asked directly "what happens when the [phase] task completes and the worktree is destroyed? does it merge back to the original task worktree?" The honest answer is no: nothing merges a child worktree/branch "back" into anything — git worktrees don't nest, and `task.complete`'s only defined target is `main`.

## Design Decisions

- **Generic lifecycle support is available:** `smaqit-extensions` Task 020 was completed and released as `v1.10.0` on 2026-07-29. It adds optional `Parent: NNN` task metadata plus parent-owned start/complete behavior. A child joins the active parent's branch/worktree and inherits its mode; child completion updates only task state. The parent alone merges and cleans up after every child is completed.
- **Dedicated feature-cycle parent:** `smaqit.feature-new` must create and start one dedicated parent task during Phase 0. The five phase tasks are children, not the lifecycle owner. The parent must remain In Progress through the whole cycle, so Phase 1 cannot itself be the parent and still be completed at the Phase 1 gate.
- **Child creation order:** Start the parent before creating phase children. `task.create --parent NNN` validates that the parent is active and writes the child task and planning state into the parent worktree; creating all five standalone tasks first would defeat the contract.
- **Per-phase completion:** Phases 1–5 still have independent status, acceptance criteria, and findings. Their `task.complete` invocations are child exits and must not merge, delete the shared branch, remove the worktree, or rebuild the workspace. Complete the parent only after Phase 5 and all child tasks are Completed.
- **Release PR is the sole vehicle for landing remaining changes — parent `task.complete` must not merge the feature branch.** The current Phase 5 already runs the release chain (`release-analysis` → `release-approval` → `release-prepare-files` → `release-git-pr`). In the parent-owned model, this chain commits CHANGELOG + version bump to the feature branch, then creates a release PR (feature-branch → `main`, title `"Prepare release vX.Y.Z"`). The release PR carries ALL remaining changes: any Phase 4 post-deployment spec amendments, plus the version bump and changelog updates from Phase 5. If parent `task.complete` were to merge the feature branch directly into `main`, it would bypass the release PR's post-merge automation — no tag, no binary builds, no GitHub release. The parent merge step MUST be a no-op, and the release PR MUST be the vehicle.

  **Concrete end-of-cycle flow:**
  1. Phase 5 child starts (joins parent worktree).
  2. Phase 5 runs the release chain on the feature branch:
     - `smaqit.release-analysis` → determines next version from the feature branch's commit delta.
     - `smaqit.release-approval` → confirms version with user.
     - `smaqit.release-prepare-files` → commits CHANGELOG + version bump to the feature branch.
     - `smaqit.release-git-pr` → pushes the feature branch and creates/updates a PR titled `"Prepare release vX.Y.Z"` against `main`. The PR description documents the post-merge automation (tag, builds, release).
  3. Phase 5 child `task.complete` — state update only (no merge, no cleanup).
  4. **Gate:** Release PR is merged by a human. Post-merge workflow fires: creates tag, builds binaries, publishes GitHub Release.
  5. Parent `task.complete` runs:
     - All 5 children verified Completed → passes resolver gate.
     - Merge step: `git merge feature-branch` into `main` — **no-op** because the release PR already merged all feature-branch content into `main` ("Already up to date.").
     - Worktree removed, branch deleted, workspace rebuilt.
     - Parent task moves to Completed.

  **Why this works:** `release-prepare-files` commits to the feature branch; `release-git-pr` opens a PR from that same branch. When the PR merges, `main` receives all feature-branch content (development work + post-deploy amendments + release prep). The subsequent parent merge is a clean no-op. The release PR's post-merge automation fires correctly because the PR title matches the `"Prepare release vX.Y.Z"` pattern.

  **Edge case — release PR not yet merged when parent completes:** If the human has not merged the release PR yet, parent `task.complete`'s merge step would fast-forward `main` to the feature branch tip, silently bypassing the release automation (no tag, no build, no GitHub release). The `smaqit.feature-new` skill MUST ensure the release PR is merged before invoking parent `task.complete`. The Phase 5 gate documents this: "Release PR merged and post-merge automation confirmed (tag exists, binaries built, GitHub release published)." If the gate is not met, parent completion is blocked.

  **Edge case — no post-deploy amendments, no release needed:** If Phases 4–5 produce zero branch-local changes (no amendments, release not applicable), Phase 5 may skip the release chain entirely. In this case, Phase 3's PR already landed all feature work on `main`. Parent `task.complete` merge step is a no-op. Cleanup proceeds normally.

  **Edge case — repo auto-deletes head branch on PR merge:** If the target repo has "Automatically delete head branches" enabled, the release PR merge deletes the feature branch. When parent `task.complete` runs, the resolver may not find the branch. The merge step skips silently ("branch does not exist"). Worktree removal still happens (the worktree directory is local and unaffected by remote branch deletion). Document this as a Gotcha: repos with auto-delete should either disable it or accept that the parent merge step is skipped.

## Implementation Steps

All changes are confined to `skills/smaqit.feature-new/SKILL.md`. No other files are modified. The work is done in the smaqit source repo (not smaqit-extensions — the parent lifecycle contract is already shipped there).

### Step 1 — Rewrite Phase 0 (Task Creation)

Replace the current Phase 0 (which creates 5 standalone phase tasks) with a parent-first flow:

1. **Create the feature-cycle parent task:**
   ```
   task.create "Feature: <brief feature description>"
   ```
   - This gets the next available task number (e.g., 098). Record it as `$PARENT`.
   - The parent task file describes the feature at a high level. Its Acceptance Criteria are: all 5 phase children completed, amendment gate clear, release tagged (if deployed).

2. **Start the parent in Assisted mode** (the default — user gates at each phase):
   ```
   task.start $PARENT
   ```
   - This creates the feature branch (`task/$PARENT-feature-<slug>`) and its sibling worktree.
   - The parent remains In Progress for the entire feature cycle.

3. **Create 5 phase children inside the parent worktree:**
   ```
   task.create "Phase 1 — Spec Revalidation for <feature>" --parent $PARENT
   task.create "Phase 2 — Development for <feature>" --parent $PARENT
   task.create "Phase 3 — Deployment for <feature>" --parent $PARENT
   task.create "Phase 4 — Validation for <feature>" --parent $PARENT
   task.create "Phase 5 — Close-out for <feature>" --parent $PARENT
   ```
   - Each child's task file gets `**Parent:** $PARENT` and is written into the parent's worktree.
   - Each child's Acceptance Criteria are the phase-specific gates from the current skill (unchanged).
   - The children's task numbers are sequential (e.g., 099–103). Record them as `$P1`–`$P5`.

4. **Gate:** All 6 task files exist in `.smaqit/tasks/`. All 6 entries appear in the parent worktree's `PLANNING.md` (Active table, Not Started). Operator confirms mode and the task set.

### Step 2 — Rewrite Phase 1 (Spec Revalidation)

Change the first and last steps of Phase 1 from standalone `task.start`/`task.complete` to child-aware invocations:

1. **Start the Phase 1 child:**
   ```
   task.start $P1
   ```
   - The resolver detects `Parent: $PARENT`, confirms the parent is In Progress in its registered worktree, and returns the parent's branch, worktree, and mode.
   - No new branch or worktree is created.
   - The child inherits the parent's Assisted mode.
   - Message to user: `"Task $P1 joined parent Task $PARENT at <worktree-path>."`

2. **Execute spec revalidation** (unchanged from current skill — Steps 2–6 of current Phase 1).

3. **Complete the Phase 1 child:**
   ```
   task.complete $P1
   ```
   - The resolver detects the child relationship. Completion updates the child's task file (Findings populated, status → Completed, ACs checked) and moves it to Completed in the parent worktree's `PLANNING.md`.
   - **No merge, no worktree removal, no branch deletion, no workspace rebuild.**
   - The feature branch and worktree survive for Phase 2.

### Step 3 — Rewrite Phases 2–4 (Development, Deployment, Validation)

Apply the same child-aware `task.start`/`task.complete` pattern to each phase:

**Phase 2 (Development):**
```
task.start $P2    # joins parent worktree, no new branch
# ... development work (unchanged) ...
task.complete $P2 # child exit: state update only
```

**Phase 3 (Deployment):**
```
task.start $P3    # joins parent worktree, no new branch
# ... deployment work — including PR creation, human merge, CI monitoring, deploy-verify ...
task.complete $P3 # child exit: state update only
```
- **Critical:** The Deployment agent creates a **deploy PR** from the parent branch (feature-branch → `main`). The human merges this PR, and CI/CD deploys from `main`. The parent branch is NOT deleted by this merge. Phase 3's child completion does NOT merge or clean up — the parent branch survives for Phases 4–5.
- **This is NOT the release PR.** The deploy PR lands the feature code on `main` and triggers deployment. The release PR (with version bump, changelog, and post-merge automation) is created later in Phase 5. They are two distinct PRs from the same feature branch at different points in time.

**Phase 4 (Validation):**
```
task.start $P4    # joins parent worktree, no new branch
# ... validation work (unchanged) ...
task.complete $P4 # child exit: state update only
```

### Step 4 — Rewrite Phase 5 (Close-out) + Parent Completion

Phase 5 is the last child. It runs the release chain on the feature branch, creating a release PR that carries all remaining changes. The parent completes only after the release PR is merged.

1. **Start the Phase 5 child:**
   ```
   task.start $P5    # joins parent worktree, inherits parent mode
   ```

2. **Confirm all prior phase tasks (1–4) are closed** in the parent worktree's `PLANNING.md`.

3. **Re-run the amendment scan** (belt-and-suspenders):
   ```
   bash [SMAQIT_SKILLS_DIR]/smaqit.feature-new/scripts/check-amendments.sh specs/
   ```
   If no matches, skip the review step.

4. **Run the release chain on the feature branch:**
   - `smaqit.release-analysis` → determines next version from the feature branch's commit delta since the last release marker.
   - `smaqit.release-approval` → confirms version with user.
   - `smaqit.release-prepare-files` → commits CHANGELOG + version bump to the **feature branch** (the current branch in the parent worktree).
   - `smaqit.release-git-pr` → pushes the feature branch, creates/updates a PR titled `"Prepare release vX.Y.Z"` against `main`. The PR description documents the post-merge automation.
   - **Critical:** The release PR is the sole vehicle for landing remaining changes on `main`. The feature branch MUST NOT be merged directly — doing so would bypass the post-merge automation (no tag, no binary builds, no GitHub release).

5. **Gate:** Release PR is merged by a human. Confirm the post-merge automation fired:
   - Git tag `vX.Y.Z` exists (`git tag -l "vX.Y.Z"`).
   - GitHub Release published with binaries and changelog.
   - If the gate is not met, parent completion is blocked.

6. **Complete the Phase 5 child:**
   ```
   task.complete $P5 # child exit: state update only
   ```

7. **Complete the parent task:**
   ```
   task.complete $PARENT
   ```
   - The resolver verifies all 5 children are Completed.
   - Parent Findings are populated (summary of the feature cycle).
   - **Merge step:** `git merge feature-branch` into `main` — **no-op** because the release PR already merged all feature-branch content into `main` ("Already up to date.").
   - Worktree removed, branch deleted, workspace rebuilt.
   - Parent task moves to Completed in the primary `PLANNING.md`.

   **If the release PR has NOT been merged yet:** the merge step would fast-forward `main` to the feature branch tip, silently bypassing the release automation. The Phase 5 gate (step 5 above) blocks parent completion until the release PR is confirmed merged. If the gate check fails, stop and report: "Release PR #N has not been merged. Parent completion is blocked until the release PR is merged and post-merge automation is confirmed."

8. **Edge case — no release needed:** If the feature does not warrant a release (e.g., internal tooling change, no user-facing impact), Phase 5 may skip the release chain (steps 4–5). In this case, Phase 3's deploy PR is the only PR. Parent `task.complete` merge step is a no-op (all content already on `main` from Phase 3's PR). Cleanup proceeds normally.

### Step 5 — Update Gotchas, Scope, and Examples

1. **Add Gotcha — "Phase tasks are children, not standalone":** An agent that invokes `task.start` for a phase task outside the parent context (e.g., from `main` instead of the parent worktree) will get a resolver error — the parent must be In Progress in its registered worktree. The correct entry point is always `smaqit.feature-new` Phase 0, which creates and starts the parent first.

2. **Add Gotcha — "Parent `task.complete` must not merge the feature branch — the release PR is the sole vehicle":** If parent `task.complete` runs before the release PR is merged, its merge step would fast-forward `main` to the feature branch tip, silently bypassing the post-merge release automation (no tag, no binary builds, no GitHub release). The Phase 5 gate explicitly checks that the release PR is merged and the post-merge automation fired before parent completion is allowed. If the gate fails, stop and report the blocking condition.

3. **Add Gotcha — "Repo auto-deletes head branch on PR merge":** If the target repo has "Automatically delete head branches" enabled, the release PR merge (Phase 5) deletes the feature branch. When parent `task.complete` runs, the merge step skips silently ("branch does not exist"). Worktree removal still happens (local directory, unaffected by remote branch deletion). Post-deploy amendments and release prep are already on `main` via the release PR, so nothing is lost. Document that repos with this setting work correctly — no configuration change needed.

4. **Add Gotcha — "Two PRs from the same feature branch":** Phase 3 opens a deploy PR; Phase 5 opens a release PR. Both are from the same feature branch to `main`, at different points in the branch's history. GitHub handles this correctly: the second PR shows only the commits added since the first PR was merged (Phase 4 amendments + release prep). No rebase or branch recreation is needed between PRs.

5. **Update Scope** to note that this task does NOT change the smaqit-extensions task lifecycle itself — the parent contract is already shipped in v1.10.0. This task only adopts it in `smaqit.feature-new`.

6. **Update Examples** to show the new Phase 0 parent-first creation flow, the child-aware Phase 1–5 `task.start`/`task.complete` invocations, and the two-PR flow (deploy PR in Phase 3, release PR in Phase 5).

### Step 6 — Validate

1. Read through the updated `SKILL.md` end-to-end: confirm every `task.start`/`task.complete` reference is child-aware, the Phase 3 merge semantics match the resolved design, and the Examples section is consistent.
2. Run `make prepare && make test` in `installer/` to ensure no regressions in the installer's embedded-file assertions.
3. Verify that `smaqit.task-list` (invoked from the parent worktree) shows all 6 tasks (1 parent + 5 children) with correct statuses — the children appear in the Active table until completed, then move to Completed. The parent stays In Progress until Phase 5.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] `smaqit.feature-new` Phase 0 creates a dedicated feature-cycle parent task, starts it (creating ONE branch/worktree), then creates 5 phase tasks as children via `task.create --parent <id>`
- [x] Phases 1–5 each invoke child-aware `task.start` (joins parent worktree, no new branch) and `task.complete` (state update only, no merge/cleanup)
- [x] Phase 5 runs the release chain on the feature branch: `release-prepare-files` commits to the feature branch, `release-git-pr` creates a release PR (`"Prepare release vX.Y.Z"`) that is the **sole vehicle** for landing remaining changes on `main`
- [x] Phase 5 gate confirms the release PR is merged and post-merge automation fired (tag exists, binaries built, GitHub release published) before parent completion is allowed
- [x] Parent `task.complete` merge step is a no-op because the release PR already landed all feature-branch content on `main`; worktree removal and branch deletion proceed normally
- [x] If the release PR has NOT been merged, parent completion is BLOCKED with a clear message — the direct merge would bypass post-merge release automation
- [x] The skill's own text no longer contradicts itself — Phase 3's "the feature branch" (singular) is the parent's branch; two distinct PRs (deploy in Phase 3, release in Phase 5) are created from it at different points
- [x] New Gotchas document: (a) child tasks require parent context, (b) parent merge must not bypass the release PR, (c) auto-delete-head-branch on PR merge is handled correctly, (d) two PRs from the same feature branch work correctly
- [ ] ~~A Gotcha references the downstream project's task 022 as the concrete case that surfaced this design gap~~ — superseded by CONTRIBUTING.md's later "never name downstream projects in shipped skill documentation" rule (added in the task 097 sanitization session, history 070); adding it now would violate that rule. Not implemented, by design.
- [x] The standalone task workflow (`task.start`/`task.complete` without `--parent`) is unchanged — this task only adds parent adoption to `smaqit.feature-new`

## Findings

**Implementation approach:**
- Rewrote `skills/smaqit.feature-new/SKILL.md` (v2.0.0) to adopt the parent-owned subtask lifecycle shipped in `smaqit-extensions` v1.10.0 (`Parent: NNN` metadata, child-aware `task.start`/`task.complete`, resolver script) instead of building any new mechanics in smaqit itself.
- Phase 0 now creates and starts one feature-cycle parent task, then creates all 5 phase tasks as children via `task.create --parent $PARENT`; Phases 1–4 use child-aware start/complete (join parent worktree, state-update-only completion); Phase 5 runs the release chain on the shared feature branch and opens a `"Prepare release vX.Y.Z"` PR as the sole vehicle onto `main`, gated on that PR being merged before parent completion is allowed.

**Decisions made:**
- The task-022 Gotcha (AC9) is intentionally left unimplemented — it would name a downstream project's task ID in a shipped skill file, which CONTRIBUTING.md now explicitly forbids (a rule added after this task was filed). The originating incident is preserved unattributed in this task's own Notes section instead.
- Completed without going through `task.start`/branch/worktree mechanics: the implementation was already merged directly to `main` in commit `2bec633` (2026-07-29, bundled with unrelated session changes), so the task was still `Not Started` in its own file despite the work being done. The lifecycle resolver correctly refused `task.complete` on a `Not Started` task (`Task 095 must be In Progress in a registered worktree before completion`) — there is no branch/worktree to merge or clean up, so this completion updates task bookkeeping only, following the same precedent as the prior task 097 (history 070).

**Blockers encountered:**
- Lifecycle resolver blocked standard completion since the task was never started through `task.start`. Resolved by completing bookkeeping directly rather than retroactively starting a task for already-merged work.

**Follow-up identified:**
- None.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.feature-new/SKILL.md` | Modify (Phase 0/1/2/3/4 — exact sections depend on Design Decisions) |

## Notes

Discovered on a downstream project during task 022's planning (2026-07-27), while retrofitting a standalone task into a `feature-new`-style cycle after the user asked whether `feature-new` had actually been used. The user's own question — "what happens when the [phase] task completes and the worktree is destroyed? does it merge back to the original task worktree?" — is the exact right question, and the honest answer (no, nothing merges back; git worktrees don't nest) is what surfaced this as a real, previously-undocumented design gap rather than a misunderstanding on the user's part.

Related: `smaqit-extensions` Task 018 ("Converge Worktree-Aware Task Lifecycle") built the uniform per-task lifecycle. Its successor, Task 020 ("Add Parent-Owned Subtask Worktree Lifecycle"), delivered the generic parent/child contract and was released in `smaqit-extensions v1.10.0` on 2026-07-29. This task now owns only the feature-workflow adoption and its remaining Phase 3 merge-timing design decision.
