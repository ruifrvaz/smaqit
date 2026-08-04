---
name: smaqit.feature-new
description: "Use when adding a post-MVP feature to a project that has already completed a `smaqit.new-greenfield-project` run (or equivalent) and has a deployed target. Applies greenfield's task-per-phase discipline and amendment gate to iterative feature work — spec revalidation, development, deployment, validation, close-out — without requirements extraction, from-scratch specs, or a dev-VM sweep. Defaults deployment to the existing target instead of provisioning a new VM. Also use when the user says 'add a feature to this project', 'iterate on the deployed app', or asks for post-MVP work and the project already has an Infrastructure spec with `status: deployed`."
metadata:
  version: "2.0.0"
---

# Post-MVP Feature Workflow

## Steps

### Pre-conditions

All items below must be satisfied before starting. When re-entering at a later phase, confirm only the items for that phase and all earlier phases.

**Always required**
- [ ] Project has already completed at least one MVP cycle — an Infrastructure spec exists under `specs/infrastructure/`
- [ ] `gh` CLI authenticated (`gh auth login`)
- [ ] Feature requirements described in session context (no `assets/raw/` sweep — this is not a from-scratch requirements extraction)

**Required before Phase 3 (Deployment)**
- [ ] CI/CD workflows exist (`.github/workflows/deploy.yml` at minimum)
- [ ] Local Vault initialised and running with this project's existing credential paths populated (`smaqit.infrastructure-vault-loader` already completed during the original MVP cycle — re-run only if credentials have rotated or expired)

If no Infrastructure spec exists yet under `specs/infrastructure/` — this project has not been through an MVP cycle — stop and flag to the user that `smaqit.new-greenfield-project` is the correct skill instead. Do not proceed silently.

---

### Phase 0 — Task Creation (Entry Point)

The operator triggers this phase manually and sets execution mode before any work begins. Phase 0 creates a dedicated feature-cycle parent task whose branch and worktree are shared by all five phase children.

1. Decide execution mode:
   - **Assisted** — operator is present at each gate; phases do not advance without explicit approval.
   - **Autonomous** — all phases run sequentially without gate interruptions; operator reviews at Phase 5.
2. **Create the feature-cycle parent task:**
   ```
   smaqit.task-create "Feature: <brief feature description>"
   ```
   - Record the returned task number as `$PARENT`.
   - The parent task file describes the feature at a high level. Its Acceptance Criteria are: all 5 phase children completed, amendment gate clear, release tagged (if deployed).
3. **Start the parent** (in Assisted mode by default):
   ```
   smaqit.task-start $PARENT
   ```
   - This creates the feature branch (`task/$PARENT-feature-<slug>`) and its sibling worktree.
   - The parent remains In Progress for the entire feature cycle.
4. **Create 5 phase children inside the parent worktree:**
   ```
   smaqit.task-create "Phase 1 — Spec Revalidation for <feature>" --parent $PARENT
   smaqit.task-create "Phase 2 — Development for <feature>" --parent $PARENT
   smaqit.task-create "Phase 3 — Deployment for <feature>" --parent $PARENT
   smaqit.task-create "Phase 4 — Validation for <feature>" --parent $PARENT
   smaqit.task-create "Phase 5 — Close-out for <feature>" --parent $PARENT
   ```
   - Each child's task file gets `**Parent:** $PARENT` and is written into the parent's worktree.
   - Each child's Acceptance Criteria are the phase-specific gates from this skill (unchanged).
   - Record the children's task numbers as `$P1`–`$P5`.
5. **Gate:** All 6 task files exist in `.smaqit/tasks/`. All 6 entries appear in the parent worktree's `PLANNING.md` (Active table, Not Started). Operator confirms mode and the task set.

### Phase 1 — Spec Revalidation

Phase 1 is the sole owner of incremental specification generation/revalidation for this feature. No specification agent runs more than once per feature unless Phase 1 is explicitly reopened for correction.

1. **Start the Phase 1 child task:**
   ```
   smaqit.task-start $P1
   ```
   - The resolver detects `Parent: $PARENT`, confirms the parent is In Progress in its registered worktree, and returns the parent's branch, worktree, and mode.
   - No new branch or worktree is created. The child inherits the parent's mode.
   - Message to user: `"Task $P1 joined parent Task $PARENT at <worktree-path>."`
2. Invoke `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack` → `/smaqit.infrastructure` → `/smaqit.coverage` as needed. At each, apply the Incremental Spec Updates decision table in [references/spec-lifecycle-reference.md](references/spec-lifecycle-reference.md).
3. Invoke `smaqit.design-validate` for every touched/confirmed pair, then run `smaqit design validate` and `smaqit plan --phase=develop` (no `--regen`) to confirm design readiness and scope — see the Incremental Plan Resolution table in the same reference for exact behavior.
4. For any spec needing only a status bump with no content change (e.g. re-confirming `implemented` still holds after this feature's tests pass), use `smaqit.spec-status-update` rather than re-invoking a full spec agent.
5. **Gate:** All touched specs have `status: draft` (new/updated) or their existing status confirmed still accurate; every active touched spec has a current visually reviewed same-layer design pair. User reviews and approves the touched spec/design set.
6. **Record durable spec handoff** in the Phase 1 task under `Decisions made`. List exact touched/confirmed spec paths grouped by consumer:
   - **Develop** — Business, Functional, and Stack spec paths plus their exact PlantUML design Markdown paths
   - **Deploy** — Infrastructure spec paths + Stack spec paths (for runtime context) plus their exact PlantUML design Markdown paths
   - **Validate** — Coverage spec paths + all upstream specs referenced by Coverage plus their exact PlantUML design Markdown paths
   This handoff is the single source of truth for Phases 2–4; they re-read it from the Phase 1 task file so context compaction or a clean resumed session cannot lose which specs were confirmed.
7. **Complete the Phase 1 child:**
   ```
   smaqit.task-complete $P1
   ```
   - The resolver detects the child relationship. Completion updates the child's task file (Findings populated, status → Completed, ACs checked) and moves it to Completed in the parent worktree's `PLANNING.md`.
   - **No merge, no worktree removal, no branch deletion, no workspace rebuild.** The feature branch and worktree survive for Phase 2.

### Phase 2 — Development

1. **Start the Phase 2 child task:**
   ```
   smaqit.task-start $P2
   ```
   - Joins the parent worktree. No new branch or worktree is created.
2. Re-read the Phase 1 task file. Extract the **Develop** handoff paths (Business, Functional, Stack specs).
3. Invoke `/smaqit.development` with `specification_mode: prevalidated` and the Develop handoff paths. The Development agent skips specification generation, reads the confirmed specs, consolidates, plans, and implements.
   - Explicitly instruct the agent to use the canonical `<!-- amendment: DATE — description -->` tag for any spec divergence (package mismatch, config change, structural adaptation) — not a prose blockquote.
4. **Gate:** Build passes. All this feature's acceptance criteria met. Development agent sets touched specs and linked designs to `status: implemented` under the least-advanced linked-spec rule, then reruns `smaqit design validate`.
5. **Complete the Phase 2 child:**
   ```
   smaqit.task-complete $P2
   ```
   - Child exit: state update only. No merge, no cleanup. The feature branch survives for Phase 3.

### Phase 3 — Deployment

No dev-VM sweep. Deployment is mandatory — every feature cycle pushes to production through a PR as the human approval gate. An unmerged feature PR is active work awaiting approval, not a completed feature with a deferred deployment.

This phase creates a **deploy PR** (feature-branch → `main`) to land the feature code and trigger CI/CD deployment. The release PR (with version bump, changelog, and post-merge release automation) is created separately in Phase 5.

**Preflight** (before invoking the Deployment agent):

1. **Start the Phase 3 child task:**
   ```
   smaqit.task-start $P3
   ```
   - Joins the parent worktree. No new branch or worktree is created.
2. Re-read the Phase 1 task file. Extract the **Deploy** handoff paths (Infrastructure specs + Stack specs for runtime context).
3. Read `specs/stack/platform-stack.md` — the authoritative stack declaration.
4. Read `specs/infrastructure/*.md` — determine deployment topology and target environment.
5. Resolve `provisioning_mode`:
   - If Infrastructure spec shows `status: deployed`, default to `existing-owned`. This overrides `smaqit.input-deployment`'s generic `provision` default.
   - Only fall through to `smaqit.input-deployment`'s standard elicitation if genuinely ambiguous (e.g. co-hosting on another project's VM → `existing-shared`; a second, dedicated VM that's provisioned out-of-band and will never be Terraform-managed → `existing-unmanaged`).
   - If resolution falls through to `provision` — this contradicts the Pre-conditions check; stop and flag.
6. Invoke `smaqit.input-deployment` with the resolved `provisioning_mode` to confirm execution parameters.
7. Require `.github/workflows/deploy.yml`. If absent, stop and flag.
8. **Apply the deterministic trigger decision table** by reading `.github/workflows/deploy.yml` and any sibling PR-close/dispatch workflows:
   1. If `deploy.yml` has a `push` trigger covering the PR base branch → omit any deployment marker and plan to monitor the resulting `push` run. A marker-gated dispatcher may coexist but must remain false.
   2. If there is no matching push trigger → require `deploy.yml` to support `workflow_dispatch` AND require exactly one merged-PR dispatcher targeting it with a determinable PR-body sentinel. Add that exact sentinel to the PR body and plan to monitor the dispatcher then the dispatched deploy run.
   3. If a matching push trigger coexists with an unconditional or non-marker PR dispatcher → stop. Merging would cause duplicate deployments.
   4. If triggers, target workflow, dispatcher, or sentinel are missing, multiple, dynamic, or otherwise ambiguous → stop before opening the PR and report the unsupported layout.
9. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running, unsealed, and credential paths are populated.
   - **`existing-shared`:** only `secret/apps/<app-slug>/github` is loaded. Then run `bootstrap-app-to-machine.sh <app-slug> <machine-slug>` against the target machine's already-registered `base-ssh` credential.
   - **`existing-unmanaged`:** same restricted load as `existing-shared` — only `github` is loaded. The difference: `bootstrap-app-to-machine.sh <app-slug> <machine-slug>` is typically registering the machine for the *first* time here (no prior `base-ssh`), so expect its fresh-registration branch (host/provider/owner_project prompts, new keypair, manual public-key install) rather than an already-trusted credential. Use this project's own slug for `owner_project`.
10. Invoke `smaqit.infrastructure-repo-config` to sync secrets from Vault to GitHub.
   - **`existing-shared`:** restricted mode — skips `tfstate`/`cyso`, syncs only `ssh` + `github`-derived secrets. Additionally: `gh variable set VM_HOST --body <shared-vm-ip>`.
   - **`existing-unmanaged`:** identical restricted mode and manual `VM_HOST` step as `existing-shared` — the only difference is *why* there's no Terraform output (nobody's, rather than another project's).

**Invoke Deployment agent:**

11. Invoke `/smaqit.deployment` once with:
    - `specification_mode: prevalidated`
    - `deployment_path: existing-cicd-pr`
    - The Deploy handoff paths (Infrastructure + Stack)
    - Resolved `provisioning_mode` and target context
    - Base branch, selected workflow/event, and sentinel (if dispatcher mode)

    The Deployment agent owns the entire contiguous operation: process prevalidated specs, generate/update deployment artifacts, run the amendment gate, commit and push the feature branch, create the **deploy PR** with the resolved trigger plan, pause for human merge (even in Autonomous mode), monitor the exact CI run(s), invoke `smaqit.infrastructure-deploy-verify --expected-sha <deploy-run-headSha>`, update spec frontmatter to `status: deployed`, and write `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md`.

12. Validate returned evidence from the Deployment agent — must include: PR number/URL, merge SHA/time, dispatcher run ID (if applicable), deploy run ID/event/head SHA, verification result, report path, and deployed spec paths/statuses.
13. Run the amendment gate: `bash [SMAQIT_SKILLS_DIR]/smaqit.feature-new/scripts/check-amendments.sh specs/`. If matches found, review each against the Phase 2 task and confirm all are resolved. An unresolved amendment blocks Phase 3 completion.
14. **Gate:** Amendment gate clear. CI/CD run completed successfully. `deploy-verify` reports all checks PASS.
15. **Complete the Phase 3 child:**
    ```
    smaqit.task-complete $P3
    ```
    - Child exit: state update only. No merge, no cleanup.
    - **The feature branch survives the deploy PR merge.** Phase 3's child completion does NOT merge the feature branch into `main` — the deploy PR already handled that. The feature branch remains the live cycle branch for Phases 4–5.
    - Ensure any amendments are captured under `Decisions made`.

### Phase 4 — Validation

1. **Start the Phase 4 child task:**
   ```
   smaqit.task-start $P4
   ```
   - Joins the parent worktree. No new branch or worktree is created.
2. Re-read the Phase 1 task file. Extract the **Validate** handoff paths (Coverage specs + all upstream specs referenced by Coverage).
3. Invoke `/smaqit.validation` with `specification_mode: prevalidated` and the Validate handoff paths. The Validation agent skips coverage specification generation, reads the confirmed specs, consolidates, generates test artifacts, and executes tests against the deployed system.
4. If any spec is found inconsistent with the live system: amend in-place on the feature branch with the canonical `amendment:` annotation. These post-deployment amendments will be carried to `main` by the Phase 5 release PR.
5. **Gate:** All validation checks pass. User signs off.
6. **Complete the Phase 4 child:**
   ```
   smaqit.task-complete $P4
   ```
   - Child exit: state update only. No merge, no cleanup. The feature branch survives for Phase 5, carrying any post-deployment spec amendments.

### Phase 5 — Close-out + Parent Completion

Phase 5 is the last child task. It runs the release chain on the feature branch, creating a **release PR** that is the sole vehicle for landing any remaining changes (post-deployment spec amendments from Phase 4, plus version bump and changelog) on `main`. The parent task completes only after the release PR is merged and post-merge automation is confirmed.

1. **Start the Phase 5 child task:**
   ```
   smaqit.task-start $P5
   ```
   - Joins the parent worktree. No new branch or worktree is created.

2. **Confirm all prior phase tasks (1–4) are closed** in the parent worktree's `PLANNING.md`. Deployment is mandatory — Phase 3 must be completed, not deferred.

3. **Re-run the amendment scan** (belt-and-suspenders):
   ```
   bash [SMAQIT_SKILLS_DIR]/smaqit.feature-new/scripts/check-amendments.sh specs/
   ```
   If no matches, skip the review step.

4. **Run the release chain on the feature branch.** The feature branch (current branch in the parent worktree) carries all work from Phases 1–4 plus any post-deployment amendments written in Phase 4.

   a. `smaqit.release-analysis` — determines the next version from the commit delta since the last release marker.
   b. `smaqit.release-approval` — confirms the version with the user.
   c. `smaqit.release-prepare-files` — commits CHANGELOG + version bump to the **feature branch**.
   d. `smaqit.release-git-pr` — pushes the feature branch and creates/updates a PR against `main` titled `"Prepare release vX.Y.Z"`. The PR description must document the post-merge automation (tag creation, binary builds, GitHub release).

   **Critical:** The release PR is the **sole vehicle** for landing remaining changes on `main`. The feature branch MUST NOT be merged directly by parent `task.complete` — doing so would bypass the post-merge release automation (no tag, no binary builds, no GitHub release).

5. **Gate:** Release PR is merged by a human. Confirm the post-merge automation fired:
   - Git tag `vX.Y.Z` exists: `git tag -l "vX.Y.Z"`
   - GitHub Release published with binaries and changelog.
   - **If the release PR has NOT been merged yet, parent completion is BLOCKED.** Stop and report: `"Release PR #N has not been merged. Parent completion is blocked until the release PR is merged and post-merge automation is confirmed."`

6. **Complete the Phase 5 child:**
   ```
   smaqit.task-complete $P5
   ```
   - Child exit: state update only. No merge, no cleanup.

7. **Complete the parent task:**
   ```
   smaqit.task-complete $PARENT
   ```
   - The resolver verifies all 5 children are Completed. Blocks if any child is still In Progress or Not Started.
   - Parent Findings are populated with a summary of the feature cycle (not per-phase detail).
   - **Merge step:** `git merge feature-branch` into `main` — **no-op** because the release PR already merged all feature-branch content into `main` ("Already up to date.").
   - Worktree is removed, branch is deleted, workspace is rebuilt.
   - Parent task moves to Completed in the primary `PLANNING.md`.

8. **Edge case — no release needed:** If the feature does not warrant a release (e.g., internal tooling change, no user-facing impact), skip the release chain (steps 4–5). Phase 3's deploy PR is the only PR. Parent `task.complete` merge step is still a no-op (all content already on `main` from Phase 3's PR). Cleanup proceeds normally.

---

## Output

- Feature implemented, specs updated in place (not regenerated from scratch)
- Deployed to the existing target via a deploy PR (Phase 3)
- Remaining changes (post-deployment amendments, version bump, changelog) landed on `main` via a release PR (Phase 5)
- No unresolved `amendment:` annotations at close-out
- All 5 phase children and the feature-cycle parent closed in `PLANNING.md`
- One feature branch and worktree shared across the entire cycle, cleaned up at parent completion

## Scope

- Covers post-MVP iterative feature work on a project that already has a deployed target. Requires an existing Infrastructure spec with `status: deployed` (or resolves `provisioning_mode` to `existing-shared` for a co-hosted target another project manages, or `existing-unmanaged` for a dedicated target nobody's Terraform manages).
- Does NOT cover greenfield project setup — no requirements extraction, no from-scratch 5-layer spec generation, no dev-VM sweep. Use `smaqit.new-greenfield-project` for that.
- Does NOT retrofit `smaqit.new-greenfield-project` itself to invoke this skill for its own post-Phase-8 iterative work — that is a sensible future task once this skill is proven, not in scope here.
- Does NOT build an automated "is this project past MVP" detector. The Pre-conditions check (Infrastructure spec with a deployed target) and the `provisioning_mode` resolution fallback (flag to the user if it resolves to `provision`) are the only maturity signals used.
- Does NOT modify the `smaqit-extensions` task lifecycle contract. The parent-owned subtask lifecycle (`Parent: NNN`, child-aware `task.start`/`task.complete`, resolver script) is already shipped in `smaqit-extensions v1.10.0`. This skill only adopts it.

## Gotchas

- **Phase tasks are children, not standalone.** An agent that invokes `task.start` for a phase task outside the parent context (e.g., from `main` instead of the parent worktree) will get a resolver error — the parent must be In Progress in its registered worktree. The correct entry point is always `smaqit.feature-new` Phase 0, which creates and starts the parent first. If you find yourself on `main` and need to resume a phase, re-read Phase 0 to understand why the parent worktree is the only valid working directory for phase tasks.
- **Parent `task.complete` must not merge the feature branch — the release PR is the sole vehicle.** If parent `task.complete` runs before the Phase 5 release PR is merged, its merge step would fast-forward `main` to the feature branch tip, silently bypassing the post-merge release automation (no tag, no binary builds, no GitHub release). The Phase 5 gate explicitly checks that the release PR is merged and the post-merge automation fired before parent completion is allowed. If the gate fails, stop and report the blocking condition.
- **Two PRs from the same feature branch, at different points in history.** Phase 3 opens a deploy PR; Phase 5 opens a release PR. Both are from the same feature branch to `main`. GitHub handles this correctly: the second PR shows only the commits added since the first PR was merged (Phase 4 amendments + release prep). No rebase or branch recreation is needed between PRs. The feature branch survives both PR merges — it is only deleted at parent cleanup.
- **Repo auto-deletes head branch on PR merge — handled correctly.** If the target repo has "Automatically delete head branches" enabled, the release PR merge (Phase 5) deletes the feature branch remotely. When parent `task.complete` runs, the merge step skips silently ("branch does not exist"). Worktree removal still happens (local directory, unaffected by remote branch deletion). Post-deploy amendments and release prep are already on `main` via the release PR, so nothing is lost. No configuration change needed.
- **`check-amendments.sh` matches a bare substring, not the canonical tag format.** The script (`grep -rl "amendment:" <dir>`) is case-sensitive and matches the literal substring `amendment:` anywhere in a file — it does not require the full `<!-- amendment: DATE — description -->` HTML-comment form. A note using different wording or capitalization (e.g. `**Amendment (...)**`) will not match. Phase 2 instructs the canonical lowercase tag specifically so the gate fires.
- **`provisioning_mode` default-override lives in this skill, not in `smaqit.input-deployment`.** `smaqit.input-deployment`'s own default of `provision` is correct for its primary caller (`smaqit.new-greenfield-project`'s brand-new-project case); do not modify that skill. The override — defaulting to `existing-owned` when an Infrastructure spec already shows a deployed target — is this skill's own Phase 3 Step 1 logic.
- **No dev-VM sweep, by design.** A dev-environment validation pass makes sense for a brand-new deploy pipeline being proven for the first time; it is wasted cloud spend and time for the Nth feature landing on a pipeline already proven. If `provisioning_mode` ever resolves to `provision` here, that is itself a signal this project isn't actually past MVP — stop and flag rather than silently provisioning.
- **Amendment gate runs every time, deploy-now or defer.** Deferred deploys still run the gate at Phase 3 — deferring the push must not mean deferring visibility into unresolved spec divergence.
- **Context collapse / phase re-read** — in long sessions the conversation is summarised by the model, and summaries capture phase names and outcomes but not the exact tool calls each phase requires. At every phase boundary, re-read this SKILL.md for the upcoming phase before executing any step rather than relying on a conversation summary.

## Examples

**Input:** Project has a deployed MVP (Infrastructure spec `status: deployed`). User asks to add an Identity & Access accounts/login feature.

**Output:**
- **Phase 0:** Creates parent task 098 ("Feature: Identity & Access"), starts it (creates branch `task/098-feature-identity-access` + worktree), then creates 5 children (099–103) via `task.create --parent 098`.
- **Phase 1:** Child task 099 starts (joins parent worktree). Spec agents update/confirm the touched specs. Scope confirmed via `smaqit plan --phase=develop`. Durable handoff recorded. Child 099 completes (state update only).
- **Phase 2:** Child task 100 starts (joins parent worktree). Development agent implements with any divergence recorded via canonical `amendment:` tag. Child 100 completes (state update only).
- **Phase 3:** Child task 101 starts (joins parent worktree). `provisioning_mode` resolves to `existing-owned`. Deployment agent creates a **deploy PR** from the feature branch; human merges; CI/CD deploys and verifies. Amendment gate runs and clears. Child 101 completes (state update only). Feature branch survives.
- **Phase 4:** Child task 102 starts (joins parent worktree). Validation agent runs tests against the deployed system. Any live-system divergence is amended on the feature branch. Child 102 completes (state update only).
- **Phase 5:** Child task 103 starts (joins parent worktree). Amendment re-scan is clear. Release chain runs on the feature branch: `release-analysis` → `release-approval` → `release-prepare-files` (commits CHANGELOG + version bump) → `release-git-pr` (creates release PR `"Prepare release v1.12.0"`). Human merges release PR; post-merge automation fires (tag, builds, release). Child 103 completes (state update only).
- **Parent completion:** `task.complete 098` — all 5 children verified Completed. Merge step is a no-op (release PR already landed everything). Worktree removed, branch deleted, workspace rebuilt. Parent moves to Completed.
- **Result:** One feature branch. One parent task. Five phase children. Two PRs (deploy in Phase 3, release in Phase 5). Zero redundant branches or worktrees.

## Completion

- [ ] Phase 0: parent task created and started (branch + worktree active), 5 phase children created via `--parent`, execution mode confirmed
- [ ] Phase 1: touched specs and design pairs revalidated and visually reviewed by their specification agents, exact PlantUML Markdown paths recorded, scope confirmed via `smaqit plan --phase=develop`, durable handoff recorded, child completed
- [ ] Phase 2: implementation complete, specs set to `implemented`, any divergence recorded via canonical `amendment:` tag, child completed
- [ ] Phase 3: `provisioning_mode` resolved with the existing-target-first override; deploy PR merged, CI/CD deploy verified; amendment gate clear; child completed
- [ ] Phase 4: validation complete, any post-deployment amendments written to feature branch, child completed
- [ ] Phase 5: amendment re-scan clear; release chain run on feature branch; release PR merged and post-merge automation confirmed; child completed
- [ ] Parent: all 5 children Completed; merge step no-op (release PR landed everything); worktree removed, branch deleted, workspace rebuilt; parent moved to Completed

## Failure Handling

| Situation | Action |
|-----------|--------|
| No Infrastructure spec with `status: deployed` exists (Pre-conditions) | Stop. Flag that `smaqit.new-greenfield-project` is the correct skill instead. |
| `provisioning_mode` resolves to `provision` despite Pre-conditions passing | Stop. Flag to the user before provisioning a new VM. |
| Amendment gate reports unresolved matches | Stop at the Phase 3 gate. Resolve or explicitly accept each annotation before continuing. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark the Phase 3 task as complete. |
| Release PR has not been merged when parent completion is attempted | Stop. Report: "Release PR #N has not been merged. Parent completion is blocked until the release PR is merged and post-merge automation is confirmed." |
| Parent resolver detects a child still In Progress or Not Started | Stop. Report which children are blocking. Complete or abandon them before retrying. |
| `task.start` invoked for a phase child outside the parent context | Resolver returns an error. Re-enter from Phase 0 or confirm the parent worktree is the active working directory. |
| Spec agent returns incomplete output | Re-run with additional context or user clarification. Do not advance with incomplete specs. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification. |
