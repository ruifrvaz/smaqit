---
name: smaqit.feature-new
description: Use when adding a post-MVP feature to a project that has already completed a `smaqit.new-greenfield-project` run (or equivalent) and has a deployed target. Applies greenfield's task-per-phase discipline and amendment gate to iterative feature work — spec revalidation, development, deployment, validation, close-out — without requirements extraction, from-scratch specs, or a dev-VM sweep. Defaults deployment to the existing target instead of provisioning a new VM. Also use when the user says "add a feature to this project", "iterate on the deployed app", or asks for post-MVP work and the project already has an Infrastructure spec with `status: deployed`.
metadata:
  version: "1.0.0"
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

The operator triggers this phase manually and sets execution mode before any work begins.

1. Decide execution mode:
   - **Assisted** — operator is present at each gate; phases do not advance without explicit approval.
   - **Autonomous** — all phases run sequentially without gate interruptions; operator reviews at Phase 5.
2. Invoke `smaqit.task-create` once for each of Phases 1–5 (Spec Revalidation, Development, Deployment, Validation, Close-out).
3. **Gate:** All task files created in `.smaqit/tasks/`. Operator confirms mode and approves the task set.

### Phase 1 — Spec Revalidation

Phase 1 is the sole owner of incremental specification generation/revalidation for this feature. No specification agent runs more than once per feature unless Phase 1 is explicitly reopened for correction.

1. Invoke `smaqit.task-start` for the Phase 1 task.
2. Invoke `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack` → `/smaqit.infrastructure` → `/smaqit.coverage` as needed. At each, apply the Incremental Spec Updates decision table in [references/spec-lifecycle-reference.md](references/spec-lifecycle-reference.md).
3. Run `smaqit plan --phase=develop` (no `--regen`) to confirm scope — see the Incremental Plan Resolution table in the same reference for exact behavior.
4. For any spec needing only a status bump with no content change (e.g. re-confirming `implemented` still holds after this feature's tests pass), use `smaqit.spec-status-update` rather than re-invoking a full spec agent.
5. **Gate:** All touched specs have `status: draft` (new/updated) or their existing status confirmed still accurate. User reviews and approves the touched spec set.
6. **Record durable spec handoff** in the Phase 1 task under `Decisions made`. List exact touched/confirmed spec paths grouped by consumer:
   - **Develop** — Business, Functional, and Stack spec paths
   - **Deploy** — Infrastructure spec paths + Stack spec paths (for runtime context)
   - **Validate** — Coverage spec paths + all upstream specs referenced by Coverage
   This handoff is the single source of truth for Phases 2–4; they re-read it from the Phase 1 task file so context compaction or a clean resumed session cannot lose which specs were confirmed.
7. Invoke `smaqit.task-complete` for the Phase 1 task.

### Phase 2 — Development

1. Invoke `smaqit.task-start` for the Phase 2 task.
2. Re-read the Phase 1 task file. Extract the **Develop** handoff paths (Business, Functional, Stack specs).
3. Invoke `/smaqit.development` with `specification_mode: prevalidated` and the Develop handoff paths. The Development agent skips specification generation, reads the confirmed specs, consolidates, plans, and implements.
   - Explicitly instruct the agent to use the canonical `<!-- amendment: DATE — description -->` tag for any spec divergence (package mismatch, config change, structural adaptation) — not a prose blockquote.
4. **Gate:** Build passes. All this feature's acceptance criteria met. Development agent sets touched specs to `status: implemented`.
5. Invoke `smaqit.task-complete` for the Phase 2 task, ensuring any amendments are captured under `Decisions made`.

### Phase 3 — Deployment

No dev-VM sweep. Deployment is mandatory — every feature cycle pushes to production through a PR as the human approval gate. An unmerged feature PR is active work awaiting approval, not a completed feature with a deferred deployment.

**Preflight** (before invoking the Deployment agent):

1. Invoke `smaqit.task-start` for the Phase 3 task.
2. Re-read the Phase 1 task file. Extract the **Deploy** handoff paths (Infrastructure specs + Stack specs for runtime context).
3. Read `specs/stack/platform-stack.md` — the authoritative stack declaration.
4. Read `specs/infrastructure/*.md` — determine deployment topology and target environment.
5. Resolve `provisioning_mode`:
   - If Infrastructure spec shows `status: deployed`, default to `existing-owned`. This overrides `smaqit.input-deployment`'s generic `provision` default.
   - Only fall through to `smaqit.input-deployment`'s standard elicitation if genuinely ambiguous (e.g. co-hosting on another project's VM → `existing-shared`).
   - If resolution falls through to `provision` — this contradicts the Pre-conditions check; stop and flag.
6. Invoke `smaqit.input-deployment` with the resolved `provisioning_mode` to confirm execution parameters.
7. Require `.github/workflows/deploy.yml`. If absent, stop and flag.
8. **Apply the deterministic trigger decision table** by reading `.github/workflows/deploy.yml` and any sibling PR-close/dispatch workflows:
   1. If `deploy.yml` has a `push` trigger covering the PR base branch → omit any deployment marker and plan to monitor the resulting `push` run. A marker-gated dispatcher may coexist but must remain false.
   2. If there is no matching push trigger → require `deploy.yml` to support `workflow_dispatch` AND require exactly one merged-PR dispatcher targeting it with a determinable PR-body sentinel. Add that exact sentinel to the PR body and plan to monitor the dispatcher then the dispatched deploy run.
   3. If a matching push trigger coexists with an unconditional or non-marker PR dispatcher → stop. Merging would cause duplicate deployments.
   4. If triggers, target workflow, dispatcher, or sentinel are missing, multiple, dynamic, or otherwise ambiguous → stop before opening the PR and report the unsupported layout.
9. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running, unsealed, and credential paths are populated.
   - **`existing-shared`:** only `secret/apps/<app-slug>/github` is loaded. Then run `bootstrap-app-to-machine.sh <app-slug> <machine-slug>`.
10. Invoke `smaqit.infrastructure-repo-config` to sync secrets from Vault to GitHub.
   - **`existing-shared`:** restricted mode — skips `tfstate`/`cyso`, syncs only `ssh` + `github`-derived secrets. Additionally: `gh variable set VM_HOST --body <shared-vm-ip>`.

**Invoke Deployment agent:**

11. Invoke `/smaqit.deployment` once with:
    - `specification_mode: prevalidated`
    - `deployment_path: existing-cicd-pr`
    - The Deploy handoff paths (Infrastructure + Stack)
    - Resolved `provisioning_mode` and target context
    - Base branch, selected workflow/event, and sentinel (if dispatcher mode)

    The Deployment agent owns the entire contiguous operation: process prevalidated specs, generate/update deployment artifacts, run the amendment gate, commit and push the feature branch, create the PR with the resolved trigger plan, pause for human merge (even in Autonomous mode), monitor the exact CI run(s), invoke `smaqit.infrastructure-deploy-verify --expected-sha <deploy-run-headSha>`, update spec frontmatter to `status: deployed`, and write `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md`.

12. Validate returned evidence from the Deployment agent — must include: PR number/URL, merge SHA/time, dispatcher run ID (if applicable), deploy run ID/event/head SHA, verification result, report path, and deployed spec paths/statuses. Feature New does not duplicate PR creation, CI monitoring, verification, status changes, or reporting.
13. Run the amendment gate: `bash [SMAQIT_SKILLS_DIR]/smaqit.feature-new/scripts/check-amendments.sh specs/`. If matches found, review each against the Phase 2 task and confirm all are resolved. An unresolved amendment blocks Phase 3 completion.
14. **Gate:** Amendment gate clear. CI/CD run completed successfully. `deploy-verify` reports all checks PASS.
15. Invoke `smaqit.task-complete` for the Phase 3 task, ensuring any amendments are captured under `Decisions made`.

### Phase 4 — Validation

1. Invoke `smaqit.task-start` for the Phase 4 task.
2. Re-read the Phase 1 task file. Extract the **Validate** handoff paths (Coverage specs + all upstream specs referenced by Coverage).
3. Invoke `/smaqit.validation` with `specification_mode: prevalidated` and the Validate handoff paths. The Validation agent skips coverage specification generation, reads the confirmed specs, consolidates, generates test artifacts, and executes tests against the deployed system.
4. If any spec is found inconsistent with the live system: amend in-place with the canonical `amendment:` annotation.
5. **Gate:** All validation checks pass. User signs off.
6. Invoke `smaqit.task-complete` for the Phase 4 task, ensuring any amendments are captured under `Decisions made`.

### Phase 5 — Close-out

1. Confirm all phase tasks (1–4) are closed in `PLANNING.md`. Deployment is mandatory — Phase 3 must be completed, not deferred.
2. Re-run the amendment scan (belt-and-suspenders): `bash [SMAQIT_SKILLS_DIR]/smaqit.feature-new/scripts/check-amendments.sh specs/`. If no matches, skip the review step.
3. Run the release chain since deployment is complete: `smaqit.release-analysis` → `smaqit.release-approval` → `smaqit.release-prepare-files` → exactly one of `smaqit.release-git-local` or `smaqit.release-git-pr`.
4. Invoke `smaqit.task-complete` for the Phase 5 task.

---

## Output

- Feature implemented, specs updated in place (not regenerated from scratch)
- Deployed to the existing target (deploy-now) or cleanly deferred with a recorded reason (defer)
- No unresolved `amendment:` annotations at close-out
- All phase tasks closed in `PLANNING.md`, or explicitly open-with-reason (Phase 3 under `defer`)

## Scope

- Covers post-MVP iterative feature work on a project that already has a deployed target. Requires an existing Infrastructure spec with `status: deployed` (or resolves `provisioning_mode` to `existing-shared` for a co-hosted target another project manages).
- Does NOT cover greenfield project setup — no requirements extraction, no from-scratch 5-layer spec generation, no dev-VM sweep. Use `smaqit.new-greenfield-project` for that.
- Does NOT retrofit `smaqit.new-greenfield-project` itself to invoke this skill for its own post-Phase-8 iterative work — that is a sensible future task once this skill is proven, not in scope here.
- Does NOT build an automated "is this project past MVP" detector. The Pre-conditions check (Infrastructure spec with a deployed target) and the `provisioning_mode` resolution fallback (flag to the user if it resolves to `provision`) are the only maturity signals used.

## Gotchas

- **`check-amendments.sh` matches a bare substring, not the canonical tag format.** The script (`grep -rl "amendment:" <dir>`) is case-sensitive and matches the literal substring `amendment:` anywhere in a file — it does not require the full `<!-- amendment: DATE — description -->` HTML-comment form. A note using different wording or capitalization (e.g. `**Amendment (...)**`) will not match. Phase 2 instructs the canonical lowercase tag specifically so the gate fires.
- **`provisioning_mode` default-override lives in this skill, not in `smaqit.input-deployment`.** `smaqit.input-deployment`'s own default of `provision` is correct for its primary caller (`smaqit.new-greenfield-project`'s brand-new-project case); do not modify that skill. The override — defaulting to `existing-owned` when an Infrastructure spec already shows a deployed target — is this skill's own Phase 3 Step 1 logic.
- **No dev-VM sweep, by design.** A dev-environment validation pass makes sense for a brand-new deploy pipeline being proven for the first time; it is wasted cloud spend and time for the Nth feature landing on a pipeline already proven. If `provisioning_mode` ever resolves to `provision` here, that is itself a signal this project isn't actually past MVP — stop and flag rather than silently provisioning.
- **Amendment gate runs every time, deploy-now or defer.** Deferred deploys still run the gate at Phase 3 — deferring the push must not mean deferring visibility into unresolved spec divergence.
- **`check-amendments.sh` is reused by reference, never forked.** Invoke it via `[SMAQIT_SKILLS_DIR]/smaqit.new-greenfield-project/scripts/check-amendments.sh` — the placeholder this repo's compile step (`scripts/generate-agents.py`) resolves per platform. A hand-written relative path (`../smaqit.new-greenfield-project/...`) is not how any existing cross-skill reference in this repo works and will not survive compilation.
- **Context collapse / phase re-read** — in long sessions the conversation is summarised by the model, and summaries capture phase names and outcomes but not the exact tool calls each phase requires. At every phase boundary, re-read this SKILL.md for the upcoming phase before executing any step rather than relying on a conversation summary.

## Examples

**Input:** Project has a deployed MVP (Infrastructure spec `status: deployed`). User asks to add an Identity & Access accounts/login feature, deploy-now.
**Output:** Phase 1 updates/creates the touched specs and confirms scope via `smaqit plan --phase=develop`. Phase 2 implements with any divergence (e.g. a new runtime dependency, a new application secret) recorded via the canonical `amendment:` tag. Phase 3 resolves `provisioning_mode` to `existing-owned`, runs the amendment gate (which now correctly catches the recorded divergence and blocks until it's resolved), then deploys and verifies. Phase 4 validates. Phase 5 confirms all phases closed, re-scans for amendments, and tags a release since this run deployed.

## Completion

- [ ] Phase 0: all phase tasks created, execution mode and deploy-now/defer confirmed
- [ ] Phase 1: touched specs revalidated per the Incremental Spec Updates decision table, scope confirmed via `smaqit plan --phase=develop`
- [ ] Phase 2: implementation complete, specs set to `implemented`, any divergence recorded via canonical `amendment:` tag
- [ ] Phase 3: `provisioning_mode` resolved with the existing-target-first override; amendment gate run and clear; deploy-now path verified, or defer path recorded with an open task and reason
- [ ] Phase 4: validation complete
- [ ] Phase 5: all phase tasks confirmed closed or explicitly deferred-with-reason; amendment re-scan clear; release tagged only if this run deployed

## Failure Handling

| Situation | Action |
|-----------|--------|
| No Infrastructure spec with `status: deployed` exists (Pre-conditions) | Stop. Flag that `smaqit.new-greenfield-project` is the correct skill instead. |
| `provisioning_mode` resolves to `provision` despite Pre-conditions passing | Stop. Flag to the user before provisioning a new VM. |
| Amendment gate reports unresolved matches | Stop at the Phase 3 gate. Resolve or explicitly accept each annotation before continuing, in both deploy-now and defer paths. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark the Phase 3 task as complete. |
| Spec agent returns incomplete output | Re-run with additional context or user clarification. Do not advance with incomplete specs. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification. |
